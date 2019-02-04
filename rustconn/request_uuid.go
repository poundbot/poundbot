package rustconn

import (
	"net/http"

	"context"

	uuid "github.com/satori/go.uuid"
)

type RequestUUID struct{}

func (ru RequestUUID) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			requestUUID := r.Header.Get("X-Request-ID")
			if requestUUID == "" {
				requestUUID = uuid.NewV4().String()
			}

			ctx := context.WithValue(r.Context(), contextKeyRequestUUID, requestUUID)
			r = r.WithContext(ctx)
			w.Header().Set("X-Request-ID", requestUUID)

			next.ServeHTTP(w, r)
		},
	)
}
