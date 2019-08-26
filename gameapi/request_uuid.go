package gameapi

import (
	"net/http"

	"context"

	"github.com/gofrs/uuid"
)

type requestUUID struct{}

func (ru requestUUID) handle(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			requestUUID := r.Header.Get("X-Request-ID")
			if requestUUID == "" {
				rUUID, err := uuid.NewV4()
				if err != nil {
					http.Error(w, "could not create UUID", http.StatusInternalServerError)
					return
				}
				requestUUID = rUUID.String()
			}

			ctx := context.WithValue(r.Context(), contextKeyRequestUUID, requestUUID)
			r = r.WithContext(ctx)
			w.Header().Set("X-Request-ID", requestUUID)

			next.ServeHTTP(w, r)
		},
	)
}
