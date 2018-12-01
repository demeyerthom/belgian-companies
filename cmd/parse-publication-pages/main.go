package main

import (
	"bytes"
	"context"
	"github.com/alecthomas/kingpin"
	"github.com/demeyerthom/belgian-companies/pkg/model"
	"github.com/demeyerthom/belgian-companies/pkg/parser"
	"github.com/demeyerthom/belgian-companies/pkg/util"
	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

var (
	// Resources
	appName           = "parse-publication-pages"
	rootURL           = "proxy://www.ejustice.just.fgov.be"
	reader            *kafka.Reader
	writer            *kafka.Writer
	publicationParser *parser.PublicationParser

	// Common flags
	inputTopic    = kingpin.Flag("input-topic", "the kafka input topic").Envar("INPUT_TOPIC").Default("publication-pages").String()
	outputTopic   = kingpin.Flag("output-topic", "the kafka output topic").Envar("OUTPUT_TOPIC").Default("publications").String()
	kafkaBrokers  = kingpin.Flag("brokers", "which kafka brokers to use").Envar("BROKERS").Default("localhost:9092").Strings()
	consumerID    = kingpin.Flag("consumer-id", "the group id for the consumer").Envar("CONSUMER_ID").String()
	withDocuments = kingpin.Flag("documents", "whether to fetch parser with documents").Envar("DOCUMENTS").Bool()
	documentPath  = kingpin.Flag("document-path", "the location to download the documents to").Envar("DOCUMENT_PATH").Default("/tmp").String()
)

func init() {
	kingpin.Parse()

	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)
	log.AddHook(util.NewApplicationHook(appName))

	publicationParser = parser.NewPublicationParser()

	readerConfig := kafka.ReaderConfig{
		Brokers:     *kafkaBrokers,
		Topic:       *inputTopic,
		Logger:      util.NewWrappedLogger(),
		ErrorLogger: util.NewWrappedLogger(),
	}
	if *consumerID != "" {
		readerConfig.GroupID = *consumerID
	}
	reader = kafka.NewReader(readerConfig)

	writer = kafka.NewWriter(kafka.WriterConfig{
		Brokers:  *kafkaBrokers,
		Topic:    *outputTopic,
		Balancer: &kafka.Hash{},
	})
}

func main() {
	c := make(chan os.Signal, 3)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)
	go func() {
		<-c

		writer.Close()
		log.Info("closed kafka writer")

		reader.Close()
		log.Info("closed cron")

		os.Exit(0)
	}()

	count := 0

	for {
		log.Debug("fetching next message")
		message, err := reader.FetchMessage(context.Background())
		util.Check(err)

		buf := bytes.NewBuffer(message.Value)

		publicationPage, err := model.DeserializePublicationPage(buf)
		util.Check(err)

		newPublications, err := publicationParser.ParsePublicationPage([]byte(publicationPage.Raw))
		util.Check(err)

		var messages []kafka.Message

		for _, publication := range newPublications {
			var buf bytes.Buffer
			err = publication.Serialize(&buf)
			util.Check(err)

			if *withDocuments {
				err = publicationParser.DownloadFile(*documentPath+publication.FileLocation, rootURL+publication.FileLocation)
				util.Check(err)
			}

			messages = append(messages, kafka.Message{Value: buf.Bytes()})

			count = count + 1
			log.WithField("publication", publication).Debug("writing new publication")
		}

		err = writer.WriteMessages(context.Background(), messages...)
		util.Check(err)

		if *consumerID != "" {
			err = reader.CommitMessages(context.Background(), message)
			util.Check(err)
		}
	}
}
