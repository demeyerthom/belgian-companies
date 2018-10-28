package main

import (
	"context"
	"encoding/json"
	"github.com/alecthomas/kingpin"
	"github.com/demeyerthom/belgian-companies/pkg/publications"
	"github.com/demeyerthom/belgian-companies/pkg/utils"
	"github.com/olivere/elastic"
	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
	"gopkg.in/sohlich/elogrus.v3"
)

var (
	// Resources
	appName       = "parse-publication-pages"
	rootURL       = "proxy://www.ejustice.just.fgov.be"
	reader        *kafka.Reader
	writer        *kafka.Writer
	elasticClient *elastic.Client

	// Common flags
	inputTopic      = kingpin.Flag("input-topic", "the kafka input topic").Default("publication-pages").String()
	outputTopic     = kingpin.Flag("output-topic", "the kafka output topic").Default("publications").String()
	kafkaBrokers    = kingpin.Flag("brokers", "which kafka brokers to use").Default("localhost:9092").Strings()
	elasticEndpoint = kingpin.Flag("elastic-endpoint", "the Elasticsearch endpoint").Default("http://localhost:9200").String()
	consumerID      = kingpin.Flag("consumer-group-id", "the group id for the consumer").Default("parse-publications-20181028").String()
	withDocuments   = kingpin.Flag("documents", "whether to fetch publications with documents").Bool()
	documentPath    = kingpin.Flag("document-path", "the location to download the documents to").Default("/tmp").String()
)

func init() {
	kingpin.Parse()
	var err error

	elasticClient, err = elastic.NewClient(elastic.SetURL(*elasticEndpoint))
	utils.Check(err)

	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)

	elasticHook, err := elogrus.NewElasticHook(elasticClient, "localhost", log.WarnLevel, "logs")
	utils.Check(err)

	log.AddHook(elasticHook)
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

		publicationPage := publications.FetchedPublicationPage{}
		err = json.Unmarshal(message.Value, &publicationPage)
		utils.Check(err)

		newPublications, err := publications.ParsePublicationPage([]byte(publicationPage.Raw))
		utils.Check(err)

		for _, publication := range newPublications {
			b, err := json.Marshal(publication)
			utils.Check(err)

			err = writer.WriteMessages(context.Background(), kafka.Message{Key: []byte(publication.ID), Value: b})
			utils.Check(err)

			if *withDocuments {
				err = publications.DownloadFile(*documentPath+publication.FileLocation, rootURL+publication.FileLocation)
				utils.Check(err)
			}

			count = count + 1
			log.WithField("publication", publication).WithField("count", count).Debug("writing new publication")
		}

		reader.CommitMessages(context.Background(), message)
	}
}
