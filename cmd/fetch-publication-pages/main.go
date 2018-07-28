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
	"net/http"
	"os"
	"time"
)

var (
	defaultTimeLayout = "2006-01-02"
	rows              = 1
	writer            *kafka.Writer
	client            *http.Client
	topic             = kingpin.Flag("topic", "the Kafka topic to write to").Default("publication-pages").String()
	kafkaBrokers      = kingpin.Flag("brokers", "which kafka brokers to use").Default("localhost:9092").Strings()
	from              = kingpin.Flag("from", "the duration").Default(time.Now().AddDate(0, 0, -1).Format(defaultTimeLayout)).String()
	to                = kingpin.Flag("to", "the duration").Default(time.Now().AddDate(0, 0, -1).Format(defaultTimeLayout)).String()
)

func init() {
	kingpin.Parse()

	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	writer = kafka.NewWriter(kafka.WriterConfig{
		Brokers:  *kafkaBrokers,
		Topic:    *topic,
		Balancer: &kafka.LeastBytes{},
	})

	client = internal.NewProxyClient()
}

func main() {
	finished := false
	pageCount := 1
	defer writer.Close()

	for finished == false {
		fromTime, _ := time.Parse(defaultTimeLayout, *from)
		toTime, _ := time.Parse(defaultTimeLayout, *to)
		result, err := publications.FetchPublicationsPage(client, rows, fromTime, toTime)
		internal.Check(err)

		if result == (publications.FetchedPublicationPage{}) {
			finished = true
			log.WithField("fetched_pages", pageCount).Info("finished fetching pages")
			os.Exit(0)
		}

		b, err := json.Marshal(result)
		internal.Check(err)

		err = writer.WriteMessages(
			context.Background(),
			kafka.Message{Value: b},
		)
		internal.Check(err)

		rows = rows + 30
		pageCount = pageCount + 1
		log.Debug(fmt.Sprintf("proceeding to row number %d", rows))
	}
}
