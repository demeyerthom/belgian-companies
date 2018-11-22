package main

import (
	"bytes"
	"context"
	"github.com/alecthomas/kingpin"
	"github.com/demeyerthom/belgian-companies/pkg/models"
	"github.com/demeyerthom/belgian-companies/pkg/publications"
	"github.com/demeyerthom/belgian-companies/pkg/utils"
	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
)

var (
	// Resources
	appName = "parse-publication-pages"
	rootURL = "proxy://www.ejustice.just.fgov.be"
	reader  *kafka.Reader
	writer  *kafka.Writer

	// Common flags
	inputTopic    = kingpin.Flag("input-topic", "the kafka input topic").Default("publication-pages").String()
	outputTopic   = kingpin.Flag("output-topic", "the kafka output topic").Default("publications").String()
	kafkaBrokers  = kingpin.Flag("brokers", "which kafka brokers to use").Default("localhost:9092").Strings()
	consumerID    = kingpin.Flag("consumer-group-id", "the group id for the consumer").Default("parse-publications").String()
	withDocuments = kingpin.Flag("documents", "whether to fetch publications with documents").Bool()
	documentPath  = kingpin.Flag("document-path", "the location to download the documents to").Default("/tmp").String()
)

func init() {
	kingpin.Parse()

	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)
	log.AddHook(utils.NewApplicationHook(appName))

	reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:     *kafkaBrokers,
		GroupID:     *consumerID,
		Topic:       *inputTopic,
		Logger:      utils.NewWrappedLogger(),
		ErrorLogger: utils.NewWrappedLogger(),
	})

	writer = kafka.NewWriter(kafka.WriterConfig{
		Brokers:  *kafkaBrokers,
		Topic:    *outputTopic,
		Balancer: &kafka.Hash{},
	})
}

func main() {
	count := 0

	for {
		log.Debug("fetching next message")
		message, err := reader.FetchMessage(context.Background())
		utils.Check(err)

		buf := bytes.NewBuffer(message.Value)

		publicationPage, err := models.DeserializePublicationPage(buf)
		utils.Check(err)

		newPublications, err := publications.ParsePublicationPage([]byte(publicationPage.Raw))
		utils.Check(err)

		var messages []kafka.Message

		for _, publication := range newPublications {
			var buf bytes.Buffer
			err = publication.Serialize(&buf)
			utils.Check(err)

			if *withDocuments {
				err = publications.DownloadFile(*documentPath+publication.FileLocation, rootURL+publication.FileLocation)
				utils.Check(err)
			}

			messages = append(messages, kafka.Message{Value: buf.Bytes()})

			count = count + 1
			log.WithField("publication", publication).Debug("writing new publication")
		}

		err = writer.WriteMessages(context.Background(), messages...)
		utils.Check(err)
		err = reader.CommitMessages(context.Background(), message)
		utils.Check(err)
	}
}
