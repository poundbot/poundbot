package discord

import (
	"time"

	"github.com/sirupsen/logrus"
)

type sessionOpener interface {
	Open() error
}

func connect(sess sessionOpener) {
	log.WithFields(logrus.Fields{"ssys": "connect"}).Info(
		"Connecting",
	)
	for {
		err := sess.Open()
		if err != nil {
			log.WithFields(logrus.Fields{"ssys": "connect"}).WithError(err).Warn(
				"Error connecting",
			)
			log.WithFields(logrus.Fields{"ssys": "connect"}).Warn(
				"Attempting Reconnect...",
			)
			time.Sleep(1 * time.Second)
			continue
		}

		log.WithFields(logrus.Fields{"ssys": "connect"}).Info(
			"Connected",
		)
		return

	}
}
