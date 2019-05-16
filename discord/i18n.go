package discord

import (
	"os"

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
		langDir := ""
		if os.Getenv("POUNDBOT_DIR") != "" {
			langDir = os.Getenv("POUNDBOT_DIR") + "/"
		}
		locBundle = i18n.NewBundle(language.English)
		locBundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
		locBundle.MustLoadMessageFile(langDir + "language/active.tr.toml")
		localizer = i18n.NewLocalizer(locBundle, lang)
	}

	uLocale, err := loc.DetectLocale()
	if err != nil {
		loadLang(language.English.String())

		log.WithFields(logrus.Fields{"sys": "LOG", "err": err}).Error(
			"Could not read localization. Defaulting to English.",
		)
	} else {
		loadLang(uLocale)
	}
}
