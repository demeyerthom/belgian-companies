package utils

import (
	"github.com/sirupsen/logrus"
	"log"
)

type ApplicationHook struct {
	applicationName string
}

// Creates a hook to be added to an instance of logger. This is called with
// `hook, err := NewContextHook("udp", "localhost:514", syslog.LOG_DEBUG, "")`
// `if err == nil { log.Hooks.Add(hook) }`
func NewApplicationHook(applicationName string) *ApplicationHook {
	return &ApplicationHook{applicationName}
}

func (h *ApplicationHook) Fire(entry *logrus.Entry) error {
	entry.Data["application"] = h.applicationName

	return nil
}

func (h *ApplicationHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func NewWrappedLogger() (wrappedLogger *log.Logger) {
	logger := logrus.StandardLogger()

	wrappedLogger = log.New(logger.Writer(), "", log.Lshortfile)

	return wrappedLogger
}
