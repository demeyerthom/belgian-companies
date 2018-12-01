package util

import (
	log "github.com/sirupsen/logrus"
	"os"
)

//Check does a check of the error input, and logs and exits the program when an error is encountered
func Check(e error) {
	if e != nil {
		log.WithError(e).Errorf("something went wrong: %s", e.Error())
		os.Exit(1)
	}
}
