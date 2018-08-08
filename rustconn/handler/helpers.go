package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"bitbucket.org/mrpoundsign/poundbot/types"
)

// handleError is a generic JSON HTTP error response
func handleError(w http.ResponseWriter, s string, restError types.RESTError) {
	w.WriteHeader(restError.StatusCode)
	err := json.NewEncoder(w).Encode(restError)
	if err != nil {
		log.Printf(s+"Error encoding %v, %s\n", restError, err)
	}
}

func methodNotAllowed(w http.ResponseWriter, s string) {
	handleError(w, s, types.RESTError{
		StatusCode: http.StatusMethodNotAllowed,
		Error:      "Method %s not allowed",
	})
}
