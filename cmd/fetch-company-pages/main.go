package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/alecthomas/kingpin"
	"github.com/demeyerthom/belgian-companies/internal"
	"github.com/demeyerthom/belgian-companies/internal/companies"
	"github.com/demeyerthom/belgian-companies/internal/publications"
	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
)

var (
	client           *http.Client
	writer           *kafka.Writer
	reader           *kafka.Reader
	companyPageTopic = kingpin.Flag("company-page-topic", "the kafka input topic").Default("company-pages").String()
	publicationTopic = kingpin.Flag("publication-topic", "the kafka input topic").Default("publications").String()
	kafkaBrokers     = kingpin.Flag("brokers", "which kafka brokers to use").Default("localhost:9092").Strings()
	consumerId       = kingpin.Flag("consumer-group-id", "the group id for the consumer").Default("parse-publications").String()
)

func init() {
	kingpin.Parse()

	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers: *kafkaBrokers,
		GroupID: *consumerId,
		Topic:   *publicationTopic,
	})

	writer = kafka.NewWriter(kafka.WriterConfig{
		Brokers:  *kafkaBrokers,
		Topic:    *companyPageTopic,
		Balancer: &kafka.LeastBytes{},
	})

	client = internal.NewProxyClient()
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

		publication := publications.Publication{}
		err = json.Unmarshal(m.Value, &publication)
		internal.Check(err)

		companyPage, err := companies.FetchCompanyPage(client, publication.DossierNumber)
		internal.Check(err)

		b, err := json.Marshal(companyPage)
		internal.Check(err)

		err = writer.WriteMessages(context.Background(), kafka.Message{Value: b})

		count = count + 1
	}

	log.WithField("count", count).Info(fmt.Sprintf("Finished processing queue: added %d company pages", count))
	os.Exit(0)
}
