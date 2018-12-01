package fetcher

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"math/rand"
	"net/http"
	"time"
)

type HttpClient interface {
	Get(url string) (resp *http.Response, err error)
}

type Fetcher struct {
	client   HttpClient
	maxSleep int
}

func (f *Fetcher) fetchRawPage(url string) (raw io.Reader, err error) {
	sleepTime := time.Duration(rand.Intn(f.maxSleep)) * time.Second
	log.Debugf("going to sleep: %d seconds", sleepTime/time.Second)
	time.Sleep(sleepTime)

	resp, err := f.client.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("invalid status code returned: %d", resp.StatusCode))
	}

	if err != nil {
		return nil, err
	}

	log.Debugf("fetched raw page for url: %s", url)

	return resp.Body, nil
}
