package gameapi

import (
	pblog "github.com/poundbot/poundbot/log"
	"github.com/sirupsen/logrus"
)

var log = pblog.Log.WithField("proc", "API")

func logWithRequest(requestURI string, sc serverContext) *logrus.Entry {
	return log.WithFields(
		logrus.Fields{
			"URI": requestURI,
			"rID": sc.requestUUID,
			"sID": sc.account.ID.Hex(),
		},
	)
}
