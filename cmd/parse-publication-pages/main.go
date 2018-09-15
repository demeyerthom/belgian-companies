package main

import (
	"context"
	"encoding/json"
	"github.com/alecthomas/kingpin"
	"github.com/demeyerthom/belgian-companies/pkg/errors"
	"github.com/demeyerthom/belgian-companies/pkg/publications"
	"github.com/robfig/cron"
	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

var (
	inputTopic    = kingpin.Flag("input-topic", "the kafka input topic").Envar("KAFKA_INPUT_TOPIC").Default("publication-pages").String()
	outputTopic   = kingpin.Flag("output-topic", "the kafka output topic").Envar("KAFKA_OUTPUT_TOPIC").Default("publications").String()
	kafkaBrokers  = kingpin.Flag("brokers", "which kafka brokers to use").Envar("KAFKA_BROKERS").Default("localhost:9092").Strings()
	consumerId    = kingpin.Flag("consumer-group-id", "the group id for the consumer").Envar("KAFKA_CONSUMER_ID").Default("parse-publications").String()
	withDocuments = kingpin.Flag("documents", "whether to fetch publications with documents").Envar("WITH_DOCUMENTS").Bool()
	documentPath  = kingpin.Flag("document-path", "the location to download the documents to").Envar("DOCUMENT_PATH").Default("/tmp").String()
	rootUrl       = "http://www.ejustice.just.fgov.be"
	reader        *kafka.Reader
	writer        *kafka.Writer
	crons         *cron.Cron
	logHandler    *os.File
	cronSpec      = kingpin.Flag("cron", "the cron specification to run").Envar("CRON_SPEC").Default("0 5 * * *").String()
	logFile       = kingpin.Flag("log-file", "the log file to write to").Envar("LOG_FILE").Default("/var/log/belgian-companies/parse-publication-pages.log").String()
)

func init() {
	kingpin.Parse()
	var err error

	if _, err := os.Stat(*logFile); os.IsNotExist(err) {
		os.Create(*logFile)
	}

	logHandler, err = os.OpenFile(*logFile, os.O_APPEND|os.O_WRONLY, 0600)
	errors.Check(err)

	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(logHandler)
	log.SetLevel(log.DebugLevel)

	reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers: *kafkaBrokers,
		GroupID: *consumerId,
		Topic:   *inputTopic,
	})

	writer = kafka.NewWriter(kafka.WriterConfig{
		Brokers:  *kafkaBrokers,
		Topic:    *outputTopic,
		Balancer: &kafka.LeastBytes{},
	})

	crons = cron.New()
}

func main() {
	c := make(chan os.Signal, 3)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)
	go func() {
		<-c

		writer.Close()
		log.Info("closed kafka writer")

		reader.Close()
		log.Info("closed kafka reader")

		crons.Stop()
		log.Info("closed cron")

		log.Info("terminated")
		logHandler.Close()
		os.Exit(0)
	}()

	crons.Start()
	crons.AddFunc(*cronSpec, parsePublicationPages)
	log.Info("started")

	select {}

}

func parsePublicationPages() {
	count := 0

	for {
		log.Debug("fetching next message")
		m, err := reader.ReadMessage(context.Background())
		errors.Check(err)

		if m.Value == nil {
			break
		}

		publicationPage := publications.FetchedPublicationPage{}
		err = json.Unmarshal(m.Value, &publicationPage)
		errors.Check(err)

		newPublications, err := publications.ParsePublicationPage([]byte(publicationPage.Raw))
		errors.Check(err)

		for _, publication := range newPublications {
			b, err := json.Marshal(publication)
			errors.Check(err)

			err = writer.WriteMessages(context.Background(), kafka.Message{Value: b})
			errors.Check(err)

			if *withDocuments {
				err = publications.DownloadFile(*documentPath+publication.FileLocation, rootUrl+publication.FileLocation)
				errors.Check(err)
			}

			count = count + 1
			log.WithField("publication", publication).WithField("count", count).Debug("writing new publication")
		}
	}

	log.Infof("Finished processing queue: added %d records", count)
}
