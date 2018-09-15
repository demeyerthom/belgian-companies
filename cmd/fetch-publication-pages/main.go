package main

import (
	"context"
	"encoding/json"
	"github.com/alecthomas/kingpin"
	"github.com/demeyerthom/belgian-companies/pkg/errors"
	"github.com/demeyerthom/belgian-companies/pkg/http-proxy"
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
	writer            *kafka.Writer
	client            *http.Client
	crons             *cron.Cron
	logHandler        *os.File
	topic             = kingpin.Flag("topic", "the Kafka topic to write to").Envar("KAFKA_TOPIC").Default("publication-pages").String()
	kafkaBrokers      = kingpin.Flag("brokers", "which kafka brokers to use").Envar("KAFKA_BROKERS").Default("localhost:9092").Strings()
	start             = kingpin.Flag("start", "the day from which to count").Envar("DATE_START").Default(time.Now().AddDate(0, 0, -1).Format(defaultTimeLayout)).String()
	end               = kingpin.Flag("end", "the day unti which to process").Envar("DATE_END").Default(time.Now().AddDate(0, 0, -1).Format(defaultTimeLayout)).String()
	cronSpec          = kingpin.Flag("cron", "the cron specification to run").Envar("CRON_SPEC").Default("0 4 * * *").String()
	logFile           = kingpin.Flag("log-file", "the log file to write to").Envar("LOG_FILE").Default("/var/log/belgian-companies/fetch-publication-pages.log").String()
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

	writer = kafka.NewWriter(kafka.WriterConfig{
		Brokers:  *kafkaBrokers,
		Topic:    *topic,
		Balancer: &kafka.LeastBytes{},
	})

	client = http_proxy.NewProxyClient()

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

		log.Info("terminated")
		logHandler.Close()
		os.Exit(0)
	}()

	crons.Start()
	crons.AddFunc(*cronSpec, fetchPublicationPages)
	log.Info("started")

	select {}
}

func fetchPublicationPages() {
	var defaultRows = 1
	var pageCount = 0
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
				log.Infof("finished fetching day `%s`. Fetched `%d` pages", day.Format(defaultTimeLayout), pageCount)
				break
			}

			result, err := publications.FetchPublicationsPage(client, rows, day)
			errors.Check(err)

			b, err := json.Marshal(result)
			errors.Check(err)

			err = writer.WriteMessages(context.Background(), kafka.Message{Value: b})
			errors.Check(err)

			rows = rows + 30
			pageCount = pageCount + 1
			log.Debugf("proceeding to row number %d", rows)
		}
	}

	log.Infof("finished fetching date range `%s` to `%s`. Fetched `%d` pages", *start, *end, pageCount)
}
