package internal

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
)

func Check(e error) {
	if e != nil {
		log.WithError(e).Error(fmt.Sprintf("something went wrong: %s", e.Error()))
		os.Exit(1)
	}
}
