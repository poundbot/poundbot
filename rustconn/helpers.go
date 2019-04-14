package rustconn

import (
	"encoding/json"
	"net/http"

	"github.com/poundbot/poundbot/types"
)

// handleError is a generic JSON HTTP error response
func handleError(w http.ResponseWriter, restError types.RESTError) error {
	w.WriteHeader(restError.StatusCode)
	return json.NewEncoder(w).Encode(restError)
}

func methodNotAllowed(w http.ResponseWriter) {
	handleError(w, types.RESTError{
		StatusCode: http.StatusMethodNotAllowed,
		Error:      "Method not allowed",
	})
}
