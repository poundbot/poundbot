package rustconn

import (
	"log"
	"net/http"
	"strings"

	"context"

	"github.com/blang/semver"
	"github.com/poundbot/poundbot/storage"
)

type ServerAuth struct {
	as storage.AccountsStore
}

func (sa ServerAuth) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			minVersion := semver.Version{Major: 1}
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
			// TOGO: Eventually remove this
			if game == "" {
				game = "rust"
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
