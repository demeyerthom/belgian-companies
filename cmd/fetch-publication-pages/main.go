package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/demeyerthom/belgian-companies/internal"
	"github.com/demeyerthom/belgian-companies/internal/publications"
	"github.com/segmentio/kafka-go"
	"os"
	"time"
)

var row = 1
var from = time.Now().AddDate(0, 0, -2)
var to = time.Now().AddDate(0, 0, -2)
var topic = "publication-pages"

func main() {
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{"localhost:9092"},
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	})
	defer w.Close()

	isSuccess := true

	for isSuccess {
		result, err := publications.FetchPublicationsPage(row, from, to)

		if err != nil {
			isSuccess = false
			fmt.Println("Finished fetching")
			os.Exit(0)
		}

		b, err := json.Marshal(result)
		internal.Check(err)

		err = w.WriteMessages(
			context.Background(),
			kafka.Message{Value: b},
		)
		internal.Check(err)

		row = row + 30
	}
}
