package main

import (
	"bytes"
	"context"
	"github.com/alecthomas/kingpin"
	"github.com/demeyerthom/belgian-companies/pkg/models"
	"github.com/demeyerthom/belgian-companies/pkg/utils"
	"github.com/olivere/elastic"
	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
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
	consumerId       = kingpin.Flag("consumer-group-id", "the group id for the consumer").Default("push-publications-elastic").String()
	elasticEndpoint  = kingpin.Flag("elastic-endpoint", "the Elasticsearch endpoint").Default("http://localhost:9200").String()
)

func init() {
	kingpin.Parse()
	var err error

	elasticClient, err = elastic.NewClient(elastic.SetURL(*elasticEndpoint))
	utils.Check(err)

	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)
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

		buf := bytes.NewBuffer(message.Value)
		publication, err := models.DeserializePublication(buf)
		utils.Check(err)

		if publication.DatePublication == "" {
			publication.DatePublication = "2016-01-01"
		}

		_, err = elasticClient.Index().
			Index("publications").
			Type("publication").
			BodyJson(publication).
			Do(context.Background())
		utils.Check(err)

		log.WithField("publication", publication).Debug("wrote new publication")
		reader.CommitMessages(context.Background(), message)
	}
}
