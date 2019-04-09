package rustconn

import (
	"net/http"
	"strings"

	"context"

	"bitbucket.org/mrpoundsign/poundbot/storage"
	"github.com/blang/semver"
)

type ServerAuth struct {
	as storage.AccountsStore
}

func (sa ServerAuth) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			minVersion, err := semver.Make("0.3.0")
			if err != nil {
				panic(err)
			}
			version, err := semver.Make(r.Header.Get("X-PoundBotConnector-Version"))
			if err != nil || version.LT(minVersion) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("PoundBotConnector must be updated. Please download the latest version at https://bitbucket.org/mrpoundsign/poundbot/downloads/."))
				return
			}
			s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
			if len(s) != 2 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			account, err := sa.as.GetByServerKey(s[1])
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), contextKeyServerKey, s[1])
			ctx = context.WithValue(ctx, contextKeyAccount, account)

			next.ServeHTTP(w, r.WithContext(ctx))
		},
	)
}
