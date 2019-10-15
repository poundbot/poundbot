package gameapi

import (
	"net/http"
	"strings"

	"context"

	"github.com/blang/semver"
	"github.com/poundbot/poundbot/storage"
)

type serverAuth struct {
	as storage.AccountsStore
}

func (sa serverAuth) handle(next http.Handler) http.Handler {
	minVersion := semver.Version{Major: 2}

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			version, err := semver.Make(r.Header.Get("X-PoundBotConnector-Version"))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Could not read PoundBot version. Please download the latest version at" + upgradeURL))
				return
			}
			if version.LT(minVersion) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("PoundBot must be updated. Please download the latest version at" + upgradeURL))
				return
			}

			s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
			if len(s) != 2 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			game := r.Header.Get("X-PoundBot-Game")

			if len(game) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Missing X-PoundBot-Game header."))
				return
			}

			account, err := sa.as.GetByServerKey(s[1])
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			err = sa.as.Touch(s[1])
			if err != nil {
				log.Printf("Error updating %s:%s touch", account.ID, s[1])
			}

			ctx := context.WithValue(r.Context(), contextKeyServerKey, s[1])
			ctx = context.WithValue(ctx, contextKeyAccount, account)
			ctx = context.WithValue(ctx, contextKeyGame, game)

			next.ServeHTTP(w, r.WithContext(ctx))
		},
	)
}
