package main

import (
	"bytes"
	"context"
	"github.com/alecthomas/kingpin"
	"github.com/demeyerthom/belgian-companies/pkg/fetcher"
	"github.com/demeyerthom/belgian-companies/pkg/model"
	"github.com/demeyerthom/belgian-companies/pkg/storage"
	"github.com/demeyerthom/belgian-companies/pkg/util"
	"github.com/go-redis/redis"
	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

var (
	// Resources
	appName                = "fetch-company-pages"
	reader                 *kafka.Reader
	writer                 *kafka.Writer
	companyFetcher         *fetcher.CompanyFetcher
	publicationDateStorage *storage.Storage

	// Common flags
	inputTopic   = kingpin.Flag("input-topic", "the kafka input topic").Envar("INPUT_TOPIC").Default("publications").String()
	outputTopic  = kingpin.Flag("output-topic", "the kafka output topic").Envar("OUTPUT_TOPIC").Default("company-pages").String()
	kafkaBrokers = kingpin.Flag("brokers", "which kafka brokers to use").Envar("BROKERS").Default("localhost:9092").Strings()
	consumerID   = kingpin.Flag("consumer-id", "the group id for the consumer").Envar("CONSUMER_ID").String()
	proxyUrl     = kingpin.Flag("proxy-url", "the proxy url to route the request through").Envar("PROXY_URL").Default("socks5://127.0.0.1:9150").String()
	sleep        = kingpin.Flag("sleep", "the max period to sleep after each request").Envar("SLEEP").Default("10").Int()
	redisUrl     = kingpin.Flag("redis-url", "the redis url").Envar("REDIS_URL").Default("localhost:6379").String()
)

func init() {
	kingpin.Parse()

	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)
	log.AddHook(util.NewApplicationHook(appName))

	client := util.NewTorClient(*proxyUrl)

	companyFetcher = fetcher.NewCompanyFetcher(client, *sleep)

	readerConfig := kafka.ReaderConfig{
		Brokers:     *kafkaBrokers,
		Topic:       *inputTopic,
		Logger:      util.NewWrappedLogger(),
		ErrorLogger: util.NewWrappedLogger(),
	}
	if *consumerID != "" {
		readerConfig.GroupID = *consumerID
	}
	reader = kafka.NewReader(readerConfig)

	writer = kafka.NewWriter(kafka.WriterConfig{
		Brokers:  *kafkaBrokers,
		Topic:    *outputTopic,
		Balancer: &kafka.Hash{},
	})

	redisClient := *redis.NewClient(&redis.Options{Addr: *redisUrl})
	publicationDateStorage = storage.NewStorage(storage.NewRedisAdapter(redisClient))

}

func main() {
	c := make(chan os.Signal, 3)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)
	go func() {
		<-c

		reader.Close()
		log.Info("closed kafka reader")

		writer.Close()
		log.Info("closed kafka writer")

		publicationDateStorage.Close()
		log.Info("closed redis storage")

		os.Exit(0)
	}()

	for {
		log.Debug("fetching next publication message")
		message, err := reader.FetchMessage(context.Background())
		util.Check(err)

		log.Debug("deserializing message")
		publication, err := model.DeserializePublication(bytes.NewBuffer(message.Value))
		util.Check(err)

		shouldNotProcess, err := publicationDateStorage.ShouldNotProcess(publication)
		util.Check(err)
		if shouldNotProcess {
			log.Infof("skipping fetching company pages for dossier %s", publication.DossierNumber)
			continue
		}

		log.Debug("fetching company pages")
		result, err := companyFetcher.FetchCompanyPages(publication.DossierNumber)
		util.Check(err)

		var buf bytes.Buffer
		err = result.Serialize(&buf)
		util.Check(err)

		log.Debugf("wrote new page set for company %s", publication.DossierNumber)
		err = writer.WriteMessages(context.Background(), kafka.Message{Value: buf.Bytes()})

		log.Debug("storing new publication date for company")
		err = publicationDateStorage.Update(publication)
		util.Check(err)

		log.Debug("committing message")
		err = reader.CommitMessages(context.Background(), message)
		util.Check(err)
	}
}
