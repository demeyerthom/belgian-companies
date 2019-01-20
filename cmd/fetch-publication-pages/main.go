package main

import (
	"bytes"
	"context"
	"github.com/alecthomas/kingpin"
	"github.com/demeyerthom/belgian-companies/pkg/fetcher"
	"github.com/demeyerthom/belgian-companies/pkg/util"
	"github.com/robfig/cron"
	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	// Resources
	appName            = "fetch-publication-pages"
	defaultTimeLayout  = "2006-01-02"
	writer             *kafka.Writer
	publicationFetcher *fetcher.PublicationFetcher
	cronHandler        *cron.Cron
	currentCommand     string

	// Common flags
	topic        = kingpin.Flag("topic", "the Kafka topic to write to").Envar("TOPIC").Default("publication-pages").String()
	kafkaBrokers = kingpin.Flag("brokers", "which kafka brokers to use").Envar("BROKERS").Default("localhost:9092").Strings()
	proxyUrl     = kingpin.Flag("proxy-url", "the proxy url to route the request through").Envar("PROXY_URL").Default("socks5://127.0.0.1:9150").String()
	sleep        = kingpin.Flag("sleep", "the max period to sleep after each request").Envar("SLEEP").Default("10").Int()

	// Daily runner
	cronCommandName = "cron"
	cronCommand     = kingpin.Command(cronCommandName, "run fetch publication pages on a continuous basis")
	cronSpec        = cronCommand.Flag("cron-spec", "the cron specification to run").Envar("CRON_SPEC").Default("0 0 2 * * *").String()

	// Defined range runner
	rangeCommandName = "range"
	rangeCommand     = kingpin.Command(rangeCommandName, "run fetch publication pages for a defined range")
	start            = rangeCommand.Flag("start", "the day from which to start").Envar("START_DATE").Default(time.Now().AddDate(0, 0, -1).Format(defaultTimeLayout)).String()
	end              = rangeCommand.Flag("end", "the day until which to process").Envar("END_DATE").Default(time.Now().AddDate(0, 0, -1).Format(defaultTimeLayout)).String()
)

func init() {
	currentCommand = kingpin.Parse()

	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)
	log.AddHook(util.NewApplicationHook(appName))

	writer = kafka.NewWriter(kafka.WriterConfig{
		Brokers:  *kafkaBrokers,
		Topic:    *topic,
		Balancer: &kafka.Hash{},
	})

	publicationFetcher = fetcher.NewPublicationFetcher(util.NewTorClient(*proxyUrl), *sleep)

	cronHandler = cron.New()
}

func main() {
	c := make(chan os.Signal, 3)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)
	go func() {
		<-c

		writer.Close()
		log.Info("closed kafka writer")

		cronHandler.Stop()
		log.Info("closed cron")

		os.Exit(0)
	}()

	switch currentCommand {
	case cronCommandName:
		log.Info("starting cron")
		cronHandler.AddFunc(*cronSpec, fetchPublicationPageForYesterday)
		cronHandler.Start()
		select {}
	case rangeCommandName:
		log.Info("starting range command")
		startDate, _ := time.Parse(defaultTimeLayout, *start)
		endDate, _ := time.Parse(defaultTimeLayout, *end)
		fetchPublicationPagesDateRange(startDate, endDate)
	}

}

func fetchPublicationPageForYesterday() {
	var date = time.Now().AddDate(0, 0, -1)
	fetchPublicationPagesForDay(date)
}

func fetchPublicationPagesDateRange(startDate time.Time, endDate time.Time) {
	var oneDayUnix = int64(86400)
	var startDateUnix = startDate.Unix()
	var endDateUnix = endDate.Unix()
	var pageCount = 0

	for timestamp := startDateUnix; timestamp <= endDateUnix; timestamp += oneDayUnix {
		pageCount = pageCount + fetchPublicationPagesForDay(time.Unix(timestamp, 0))
	}

	log.Infof("finished fetching date range `%s` to `%s`. Fetched `%d` pages", *start, *end, pageCount)
}

func fetchPublicationPagesForDay(day time.Time) (pageCount int) {
	var rows = 1

	for {
		result, err := publicationFetcher.FetchPublicationsPage(rows, day)
		util.Check(err)

		if result == nil {
			log.Infof("finished fetching day `%s`. Fetched `%d` pages", day.Format(defaultTimeLayout), pageCount)
			return pageCount
		}

		var buf bytes.Buffer
		err = result.Serialize(&buf)
		util.Check(err)

		err = writer.WriteMessages(context.Background(), kafka.Message{Value: buf.Bytes()})
		util.Check(err)

		rows = rows + 30
		pageCount = pageCount + 1

		log.Debugf("proceeding to row number %d", rows)
	}
}
