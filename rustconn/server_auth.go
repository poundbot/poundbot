package rustconn

import (
	"net/http"
	"strings"

	"context"

	"bitbucket.org/mrpoundsign/poundbot/storage"
	"bitbucket.org/mrpoundsign/poundbot/types"
)

type ServerAuth struct {
	as storage.AccountsStore
}

func (sa ServerAuth) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
			if len(s) != 2 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			var account types.Account

			err := sa.as.GetByServerKey(s[1], &account)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), contextKeyServerKey, s[1])
			ctx = context.WithValue(ctx, contextKeyAccount, account)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		},
	)
}
