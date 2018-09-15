package pkg

import (
	log "github.com/sirupsen/logrus"
	"os"
)

func Check(e error) {
	if e != nil {
		log.WithError(e).Errorf("something went wrong: %s", e.Error())
		os.Exit(1)
	}
}
