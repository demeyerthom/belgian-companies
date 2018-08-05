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
	"gopkg.in/mgo.v2"
	"os"
)

var (
	publicationTopic = kingpin.Flag("publication-topic", "the kafka publication topic").Default("publications").String()
	kafkaBrokers     = kingpin.Flag("brokers", "which kafka brokers to use").Default("localhost:9092").Strings()
	consumerId       = kingpin.Flag("consumer-group-id", "the group id for the consumer").Default("push-publications").String()
	mongoUrl         = kingpin.Flag("mongo-url", "the mongo database url").Default("localhost:27017").String()
	database         = kingpin.Flag("database", "the mongo database").Default("belgian-companies").String()
	collection       = kingpin.Flag("collection", " the mongo collection").Default("publications").String()
	reader           *kafka.Reader
	session          *mgo.Session
)

func init() {
	kingpin.Parse()

	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers: *kafkaBrokers,
		GroupID: *consumerId,
		Topic:   *publicationTopic,
	})

	newSession, err := mgo.Dial(*mongoUrl)
	session = newSession
	internal.Check(err)
}

func main() {
	defer reader.Close()
	defer session.Close()

	db := session.DB(*database)

	count := 0

	for {
		log.Debug("reading newpublication")
		m, err := reader.ReadMessage(context.Background())
		internal.Check(err)

		if m.Value == nil {
			break
		}

		publication := publications.Publication{}
		err = json.Unmarshal(m.Value, &publication)
		internal.Check(err)

		err = db.C(*collection).Insert(publication)
		internal.Check(err)

		count = count + 1
		log.WithField("publication", publication).WithField("count", count).Debug("wrote new publication")
	}

	log.WithField("count", count).Info(fmt.Sprintf("Finished processing queue: added %d publications", count))
	os.Exit(0)
}
