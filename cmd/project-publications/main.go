package main

import (
	"bytes"
	"context"
	"github.com/alecthomas/kingpin"
	"github.com/demeyerthom/belgian-companies/pkg/model"
	"github.com/demeyerthom/belgian-companies/pkg/util"
	"github.com/olivere/elastic"
	"github.com/olivere/elastic/config"
	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

var (
	// Resources
	appName       = "project-publications"
	reader        *kafka.Reader
	elasticClient *elastic.Client

	// Common flags
	topic        = kingpin.Flag("topic", "the kafka topic to read from").Envar("TOPIC").Default("publications").String()
	kafkaBrokers = kingpin.Flag("brokers", "which kafka brokers to use").Envar("BROKERS").Default("localhost:9092").Strings()
	consumerID   = kingpin.Flag("consumer-id", "the group id for the consumer").Envar("CONSUMER_ID").Default("publications-projector-elastic").String()
	elasticIndex = kingpin.Flag("elastic-index", "the elastic index to write to").Envar("ELASTIC_INDEX").Default("publications").String()
	elasticURL   = kingpin.Flag("elastic-url", "the elastic url").Envar("ELASTIC_URL").Default("http://127.0.0.1:9200").String()
)

func init() {
	var err error
	kingpin.Parse()

	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)
	log.AddHook(util.NewApplicationHook(appName))

	readerConfig := kafka.ReaderConfig{
		Brokers:     *kafkaBrokers,
		Topic:       *topic,
		Logger:      util.NewWrappedLogger(),
		ErrorLogger: util.NewWrappedLogger(),
	}
	if *consumerID != "" {
		readerConfig.GroupID = *consumerID
	}
	reader = kafka.NewReader(readerConfig)

	elasticConfig := config.Config{URL: *elasticURL}

	elasticClient, err = elastic.NewClientFromConfig(&elasticConfig)
	util.Check(err)
}

func main() {
	c := make(chan os.Signal, 3)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)
	go func() {
		<-c

		reader.Close()
		log.Info("closed reader")

		elasticClient.CloseIndex(*elasticIndex)
		log.Info("closed es index")

		os.Exit(0)
	}()

	for {
		log.Debug("fetching next message")
		message, err := reader.FetchMessage(context.Background())
		util.Check(err)

		buf := bytes.NewBuffer(message.Value)

		publication, err := model.DeserializePublication(buf)
		util.Check(err)

		_, err = elasticClient.Index().
			Index(*elasticIndex).
			Type("doc").
			Id(string(publication.ID)).
			BodyJson(publication).
			Refresh("wait_for").
			Do(context.Background())
		util.Check(err)

		if *consumerID != "" {
			err = reader.CommitMessages(context.Background(), message)
			util.Check(err)
		}
	}
}
