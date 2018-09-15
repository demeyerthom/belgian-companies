package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/alecthomas/kingpin"
	"github.com/demeyerthom/belgian-companies/pkg"
	"github.com/demeyerthom/belgian-companies/pkg/publications"
	"github.com/robfig/cron"
	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	defaultTimeLayout = "2006-01-02"
	defaultRows       = 1
	writer            *kafka.Writer
	client            *http.Client
	crons             *cron.Cron
	logHandler        *os.File
	topic             = kingpin.Flag("topic", "the Kafka topic to write to").Default("publication-pages").String()
	kafkaBrokers      = kingpin.Flag("brokers", "which kafka brokers to use").Default("localhost:9092").Strings()
	start             = kingpin.Flag("start", "the day from which to count").Default(time.Now().AddDate(0, 0, -1).Format(defaultTimeLayout)).String()
	end               = kingpin.Flag("end", "the day unti which to process").Default(time.Now().AddDate(0, 0, -1).Format(defaultTimeLayout)).String()
	cronSpec          = kingpin.Flag("cron", "the cron specification to run").Default("0 4 * * *").String()
	logFile           = kingpin.Flag("log-file", "the log file to write to").Default("/var/log/belgian-companies/fetch-publication-pages.log").String()
)

func init() {
	kingpin.Parse()

	var err error

	if _, err := os.Stat(*logFile); os.IsNotExist(err) {
		os.Create(*logFile)
	}

	logHandler, err = os.OpenFile(*logFile, os.O_APPEND|os.O_WRONLY, 0600)
	pkg.Check(err)

	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(logHandler)
	log.SetLevel(log.DebugLevel)

	writer = kafka.NewWriter(kafka.WriterConfig{
		Brokers:  *kafkaBrokers,
		Topic:    *topic,
		Balancer: &kafka.LeastBytes{},
	})

	client = pkg.NewProxyClient()

	crons = cron.New()
}

func main() {
	c := make(chan os.Signal, 3)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)
	go func() {
		<-c

		writer.Close()
		log.Info("closed kafka writer")

		crons.Stop()
		log.Info("closed cron")

		log.Info("fetch-publication-pages has terminated")
		logHandler.Close()
		os.Exit(0)
	}()

	crons.Start()
	crons.AddFunc(*cronSpec, fetchPublicationPages)
	log.Info("fetch-publication-pages has started")

	select {}
}

func fetchPublicationPages() {
	pageCount := 0
	startDate, _ := time.Parse(defaultTimeLayout, *start)
	endDate, _ := time.Parse(defaultTimeLayout, *end)

	oneDayUnix := int64(86400)
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
			pkg.Check(err)

			b, err := json.Marshal(result)
			pkg.Check(err)

			err = writer.WriteMessages(
				context.Background(),
				kafka.Message{Value: b},
			)
			pkg.Check(err)

			rows = rows + 30
			pageCount = pageCount + 1
			log.Debug(fmt.Sprintf("proceeding to row number %d", rows))
		}
	}

	log.Info("finished fetching date range")
}
