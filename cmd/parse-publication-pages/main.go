package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/demeyerthom/belgian-companies/internal"
	"github.com/demeyerthom/belgian-companies/internal/publications"
	"github.com/segmentio/kafka-go"
)

var inputTopic = "publication-pages"
var outputTopic = "publications"
var groupId = "publication-page-parser-2"

func main() {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		//GroupID: groupId,
		Topic: inputTopic,
	})
	defer r.Close()

	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{"localhost:9092"},
		Topic:    outputTopic,
		Balancer: &kafka.LeastBytes{},
	})
	defer w.Close()

	for {
		m, err := r.ReadMessage(context.Background())
		internal.Check(err)

		if m.Value == nil {
			fmt.Println("Done processing queue")
			break
		}

		publicationPage := publications.FetchedPublicationPage{}
		err = json.Unmarshal(m.Value, &publicationPage)
		internal.Check(err)

		newPublications, err := publications.ParsePublicationPage([]byte(publicationPage.Raw), "/tmp", false)
		internal.Check(err)

		for _, publication := range newPublications {
			fmt.Println(publication)
			b, err := json.Marshal(publication)
			internal.Check(err)

			err = w.WriteMessages(
				context.Background(),
				kafka.Message{Value: b},
			)
			internal.Check(err)
		}
	}

	fmt.Println("Finished")
}
