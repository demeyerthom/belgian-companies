package main

import (
	"context"
	"encoding/json"
	"github.com/alecthomas/kingpin"
	"github.com/demeyerthom/belgian-companies/pkg/proxy"
	"github.com/demeyerthom/belgian-companies/pkg/publications"
	"github.com/demeyerthom/belgian-companies/pkg/utils"
	"github.com/olivere/elastic"
	"github.com/robfig/cron"
	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
	"gopkg.in/sohlich/elogrus.v3"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	// Resources
	appName           = "fetch-publication-pages"
	defaultTimeLayout = "2006-01-02"
	writer            *kafka.Writer
	elasticClient     *elastic.Client
	client            *http.Client
	cronHandler       *cron.Cron
	currentCommand    string

	// Common flags
	topic           = kingpin.Flag("topic", "the Kafka topic to write to").Default("publication-pages").String()
	kafkaBrokers    = kingpin.Flag("brokers", "which kafka brokers to use").Default("localhost:9092").Strings()
	elasticEndpoint = kingpin.Flag("elastic-endpoint", "the Elasticsearch endpoint").Default("http://localhost:9200").String()
	proxyUrl        = kingpin.Flag("proxy-url", "the proxy url to route the request through").Default("socks5://127.0.0.1:9150").String()
	sleep           = kingpin.Flag("sleep", "the max period to sleep after each request").Default("10").Int()

	// Daily runner
	cronCommandName = "cron"
	cronCommand     = kingpin.Command(cronCommandName, "run fetch publication pages on a continuous basis")
	cronSpec        = cronCommand.Flag("cron-spec", "the cron specification to run").Default("0 4 * * *").String()

	// Defined range runner
	rangeCommandName = "range"
	rangeCommand     = kingpin.Command(rangeCommandName, "run fetch publication pages for a defined range")
	start            = rangeCommand.Flag("start", "the day from which to start").Default(time.Now().AddDate(0, 0, -1).Format(defaultTimeLayout)).String()
	end              = rangeCommand.Flag("end", "the day until which to process").Default(time.Now().AddDate(0, 0, -1).Format(defaultTimeLayout)).String()
)

func init() {
	currentCommand = kingpin.Parse()
	var err error

	elasticClient, err = elastic.NewClient(elastic.SetURL(*elasticEndpoint))
	utils.Check(err)

	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)

	hook, err := elogrus.NewElasticHook(elasticClient, "localhost", log.WarnLevel, "logs")
	utils.Check(err)

	log.AddHook(hook)
	log.AddHook(utils.NewApplicationHook(appName))

	writer = kafka.NewWriter(kafka.WriterConfig{
		Brokers:  *kafkaBrokers,
		Topic:    *topic,
		Balancer: &kafka.Hash{},
	})

	client = proxy.NewTorClient(*proxyUrl)

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
		cronHandler.AddFunc(*cronSpec, fetchPublicationPageForYesterday)
		cronHandler.Start()
		select {}
	case rangeCommandName:
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
		if false == publications.PublicationPageExists(client, rows, day) {
			log.Infof("finished fetching day `%s`. Fetched `%d` pages", day.Format(defaultTimeLayout), pageCount)
			return pageCount
		}

		result, err := publications.FetchPublicationsPage(client, rows, day)
		utils.Check(err)

		b, err := json.Marshal(result)
		utils.Check(err)

		err = writer.WriteMessages(context.Background(), kafka.Message{Value: b})
		utils.Check(err)

		rows = rows + 30
		pageCount = pageCount + 1

		sleepTime := time.Duration(rand.Intn(*sleep)) * time.Second
		log.Debugf("going to sleep: %d seconds", sleepTime/time.Second)
		time.Sleep(sleepTime)

		log.Debugf("proceeding to row number %d", rows)
	}
}
