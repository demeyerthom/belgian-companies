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
	defaultRows       = 1
	writer            *kafka.Writer
	client            *http.Client
	topic             = kingpin.Flag("topic", "the Kafka topic to write to").Default("publication-pages").String()
	kafkaBrokers      = kingpin.Flag("brokers", "which kafka brokers to use").Default("localhost:9092").Strings()
	start             = kingpin.Flag("start", "the day from which to count").Default(time.Now().AddDate(0, 0, -1).Format(defaultTimeLayout)).String()
	end               = kingpin.Flag("end", "the day unti which to process").Default(time.Now().AddDate(0, 0, -1).Format(defaultTimeLayout)).String()
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
	pageCount := 0
	defer writer.Close()

	startDate, _ := time.Parse(defaultTimeLayout, *start)
	endDate, _ := time.Parse(defaultTimeLayout, *end)

	oneDayUnix := int64(86400) // a day in seconds.
	startDateUnix := startDate.Unix()
	endDateUnix := endDate.Unix()

	for timestamp := startDateUnix; timestamp <= endDateUnix; timestamp += oneDayUnix {
		rows := defaultRows
		for {
			day := time.Unix(timestamp, 0)
			if false == publications.PublicationPageExists(client, rows, day) {
				log.WithField("pageCount", pageCount).WithField("date", day.Format(defaultTimeLayout)).Info("finished fetching day")
				break
			}

			result, err := publications.FetchPublicationsPage(client, rows, day)
			internal.Check(err)

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

	log.Info("finished fetching date range")
	os.Exit(0)
}
