package main

import (
	"context"
	"encoding/json"
	"github.com/alecthomas/kingpin"
	"github.com/demeyerthom/belgian-companies/pkg"
	"github.com/demeyerthom/belgian-companies/pkg/publications"
	"github.com/robfig/cron"
	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"os"
	"os/signal"
	"syscall"
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
	crons            *cron.Cron
	logHandler       *os.File
	cronSpec         = kingpin.Flag("cron", "the cron specification to run").Default("0 6 * * *").String()
	logFile          = kingpin.Flag("log-file", "the log file to write to").Default("/var/log/belgian-companies/push-publications.log").String()
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

	reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers: *kafkaBrokers,
		GroupID: *consumerId,
		Topic:   *publicationTopic,
	})

	session, err = mgo.Dial(*mongoUrl)
	pkg.Check(err)

	crons = cron.New()
}

func main() {
	c := make(chan os.Signal, 3)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)
	go func() {
		<-c

		reader.Close()
		log.Info("closed kafka reader")

		session.Close()
		log.Infof("closed mongo session")

		crons.Stop()
		log.Info("closed cron")

		log.Info("push-publications has terminated")
		logHandler.Close()
		os.Exit(0)
	}()

	crons.Start()
	crons.AddFunc(*cronSpec, pushPublications)
	log.Info("push-publications has started")

	select {}
}

func pushPublications() {
	db := session.DB(*database)

	count := 0

	for {
		log.Debug("reading new publication")
		m, err := reader.ReadMessage(context.Background())
		pkg.Check(err)

		if m.Value == nil {
			break
		}

		publication := publications.Publication{}
		err = json.Unmarshal(m.Value, &publication)
		pkg.Check(err)

		err = db.C(*collection).Insert(publication)
		pkg.Check(err)

		count = count + 1
		log.WithField("publication", publication).WithField("count", count).Debug("wrote new publication")
	}

	log.Infof("Finished processing queue: added %d publications")
}
