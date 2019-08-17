package gameapi

import (
	pblog "github.com/poundbot/poundbot/log"
	"github.com/sirupsen/logrus"
)

var log = pblog.Log.WithField("proc", "API")

func logWithRequest(sc serverContext) *logrus.Entry {
	return log.WithFields(
		logrus.Fields{
			"rID": sc.requestUUID,
			"sID": sc.account.ID.Hex(),
		},
	)
}
