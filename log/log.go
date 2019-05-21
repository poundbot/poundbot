package log

import (
	"os"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
)

var Log *logrus.Entry

type formatter struct {
	fields logrus.Fields
	lf     logrus.Formatter
}

func init() {
	l := logrus.New()
	l.SetFormatter(&nested.Formatter{
		HideKeys:    false,
		FieldsOrder: []string{"proc", "sys", "cmd", "aID", "gID", "uID", "pID"},
		NoColors:    true,
	})

	if os.Getenv("LOG_TRACE") == "on" {
		l.SetLevel(logrus.TraceLevel)
	}
	Log = l.WithField("proc", "POUNDBOT")

}
