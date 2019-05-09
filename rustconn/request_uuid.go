package rustconn

import (
	"net/http"

	"context"

	"github.com/gofrs/uuid"
)

type RequestUUID struct{}

func (ru RequestUUID) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			requestUUID := r.Header.Get("X-Request-ID")
			if requestUUID == "" {
				ruid, err := uuid.NewV4()
				if err != nil {
					http.Error(w, "could not create UUID", http.StatusInternalServerError)
					return
				}
				requestUUID = ruid.String()
			}

			ctx := context.WithValue(r.Context(), contextKeyRequestUUID, requestUUID)
			r = r.WithContext(ctx)
			w.Header().Set("X-Request-ID", requestUUID)

			next.ServeHTTP(w, r)
		},
	)
}
