package rustconn

import (
	"net/http"

	"github.com/gorilla/context"
	uuid "github.com/satori/go.uuid"
)

type RequestUUID struct{}

func (ru RequestUUID) Handle(next http.Handler) http.Handler {
	return context.ClearHandler(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				requestUUID := r.Header.Get("X-Request-ID")
				if requestUUID == "" {
					requestUUID = uuid.NewV4().String()
				}

				context.Set(r, "requestUUID", requestUUID)
				w.Header().Set("X-Request-ID", requestUUID)

				next.ServeHTTP(w, r)
			},
		),
	)
}
