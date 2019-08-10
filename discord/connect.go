package discord

import (
	"time"

	"github.com/sirupsen/logrus"
)

type sessionOpener interface {
	Open() error
}

func connect(sess sessionOpener) {
	log.WithFields(logrus.Fields{"sys": "connect"}).Info(
		"Connecting",
	)
	for {
		err := sess.Open()
		if err != nil {
			log.WithFields(logrus.Fields{"sys": "connect"}).WithError(err).Warn(
				"Error connecting; Attempting reconnect...",
			)
			time.Sleep(1 * time.Second)
			continue
		}

		log.WithFields(logrus.Fields{"sys": "connect"}).Info(
			"Connected",
		)
		return

	}
}
