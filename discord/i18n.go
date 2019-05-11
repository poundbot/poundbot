package discord

import (
	"github.com/BurntSushi/toml"
	"golang.org/x/text/language"

	loc "github.com/jmshal/go-locale"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/sirupsen/logrus"
)

var locBundle *i18n.Bundle
var localizer *i18n.Localizer

func init() {
	loadLang := func(lang string) {
		locBundle = &i18n.Bundle{DefaultLanguage: language.English}
		locBundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
		locBundle.MustLoadMessageFile("language/active.tr.toml")
		localizer = i18n.NewLocalizer(locBundle, lang)
	}

	uLocale, err := loc.DetectLocale()
	if err != nil {
		loadLang(language.English.String())

		log.WithFields(logrus.Fields{"sys": "DSCD", "ssys": "RUNNER", "err": err}).Error(
			"Could not read localization. Defaulting to English.",
		)
	} else {
		loadLang(uLocale)
	}
}
