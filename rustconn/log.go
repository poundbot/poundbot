package rustconn

import (
	"os"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
)

type formatter struct {
	fields logrus.Fields
	lf     logrus.Formatter
}

// Format satisfies the logrus.Formatter interface.
func (f *formatter) Format(e *logrus.Entry) ([]byte, error) {
	for k, v := range f.fields {
		e.Data[k] = v
	}
	return f.lf.Format(e)
}

var log = logrus.New()

func init() {
	log.SetFormatter(&formatter{
		fields: logrus.Fields{
			"proc": "API",
		},
		lf: &nested.Formatter{
			HideKeys:    false,
			FieldsOrder: []string{"proc", "sys"},
		},
	})

	if os.Getenv("LOG_TRACE") == "on" {
		log.SetLevel(logrus.TraceLevel)
	}
}
