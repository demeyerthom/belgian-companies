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
	"os"
	"os/signal"
	"syscall"
)

var (
	appName       = "push-publications"
	reader        *kafka.Reader
	elasticClient *elastic.Client

	// Common flags
	publicationTopic = kingpin.Flag("publications", "the kafka publication topic").Default("publications").String()
	kafkaBrokers     = kingpin.Flag("brokers", "which kafka brokers to use").Default("localhost:9092").Strings()
	consumerId       = kingpin.Flag("consumer-group-id", "the group id for the consumer").Default("push-publications-20181028").String()
	elasticEndpoint  = kingpin.Flag("elastic-endpoint", "the Elasticsearch endpoint").Default("http://localhost:9200").String()
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
		GroupID:     *consumerId,
		Topic:       *publicationTopic,
		Logger:      utils.NewWrappedLogger(),
		ErrorLogger: utils.NewWrappedLogger(),
	})
}

func main() {
	c := make(chan os.Signal, 3)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)
	go func() {
		<-c

		reader.Close()
		log.Info("closed kafka reader")

		os.Exit(0)
	}()

	for {
		log.Debug("reading new publication")
		message, err := reader.FetchMessage(context.Background())
		utils.Check(err)

		publication := publications.Publication{}
		err = json.Unmarshal(message.Value, &publication)
		utils.Check(err)

		_, err = elasticClient.Index().
			Index("publications").
			Type("publication").
			Id(publication.ID).
			BodyJson(publication).
			Do(context.Background())
		utils.Check(err)

		log.WithField("publication", publication).Debug("wrote new publication")
		reader.CommitMessages(context.Background(), message)
	}
}
