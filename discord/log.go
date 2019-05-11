package discord

import (
	"os"

	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func init() {
	if os.Getenv("LOG_TRACE") == "on" {
		log.SetLevel(logrus.TraceLevel)
	}
}
