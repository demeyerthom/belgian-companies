package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/alecthomas/kingpin"
	"github.com/demeyerthom/belgian-companies/internal"
	"github.com/demeyerthom/belgian-companies/internal/publications"
	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
	"os"
)

var (
	inputTopic    = kingpin.Flag("input-topic", "the kafka input topic").Default("publication-pages").String()
	outputTopic   = kingpin.Flag("output-topic", "the kafka output topic").Default("publications").String()
	kafkaBrokers  = kingpin.Flag("brokers", "which kafka brokers to use").Default("localhost:9092").Strings()
	consumerId    = kingpin.Flag("consumer-group-id", "the group id for the consumer").Default("").String()
	withDocuments = kingpin.Flag("documents", "whether to fetch publications with documents").Bool()
	documentPath  = kingpin.Flag("document-path", "the absolute location to download the documents to").Default("/tmp").String()
	rootUrl       = "http://www.ejustice.just.fgov.be"
	reader        *kafka.Reader
	writer        *kafka.Writer
)

func init() {
	kingpin.Parse()

	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
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

}

func main() {
	defer reader.Close()
	defer writer.Close()

	count := 0

	for {
		log.Debug("fetching next message")
		m, err := reader.ReadMessage(context.Background())
		internal.Check(err)

		if m.Value == nil {
			break
		}

		publicationPage := publications.FetchedPublicationPage{}
		err = json.Unmarshal(m.Value, &publicationPage)
		internal.Check(err)

		newPublications, err := publications.ParsePublicationPage([]byte(publicationPage.Raw))
		internal.Check(err)

		for _, publication := range newPublications {
			b, err := json.Marshal(publication)
			internal.Check(err)

			err = writer.WriteMessages(context.Background(), kafka.Message{Value: b})
			internal.Check(err)

			if *withDocuments {
				err = publications.DownloadFile(*documentPath+publication.FileLocation, rootUrl+publication.FileLocation)
				internal.Check(err)
			}

			count = count + 1
			log.WithField("publication", publication).WithField("count", count).Debug("writing new publication")
		}
	}

	log.WithField("count", count).Info(fmt.Sprintf("Finished processing queue: added %d records", count))
	os.Exit(0)
}
