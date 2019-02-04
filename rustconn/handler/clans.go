package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"bitbucket.org/mrpoundsign/poundbot/storage"
	"bitbucket.org/mrpoundsign/poundbot/types"
)

type clans struct {
	as     storage.AccountsStore
	logger *log.Logger
}

func NewClans(logPrefix string, as storage.AccountsStore) func(w http.ResponseWriter, r *http.Request) {
	c := clans{as: as, logger: &log.Logger{}}
	c.logger.SetPrefix(logPrefix)
	return c.Handle
}

// Handle manages clans sync HTTP requests from the Rust server
// These requests are a complete refresh of all clans
func (c *clans) Handle(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	serverKey := r.Context().Value(contextKeyServerKey).(string)
	requestUUID := r.Context().Value(contextKeyRequestUUID).(string)
	log.Printf("[%s] clansHandler: Updating all clans for %s\n", requestUUID, serverKey)

	decoder := json.NewDecoder(r.Body)
	var t []types.ServerClan
	err := decoder.Decode(&t)
	if err != nil {
		log.Println(err.Error())
		handleError(w, types.RESTError{StatusCode: http.StatusBadRequest, Error: "Could not decode clans"})
		return
	}

	clanCount := len(t)
	clans := make([]types.Clan, clanCount)
	for i, sc := range t {
		cl, err := types.ClanFromServerClan(sc)
		if err != nil {
			log.Printf("[%s] clansHandler Error: %v\n", requestUUID, err)
			handleError(w, types.RESTError{
				StatusCode: http.StatusBadRequest,
				Error:      "Error processing clan data",
			})
			return
		}
		clans[i] = *cl
	}

	err = c.as.SetClans(serverKey, clans)
	if err != nil {
		c.logger.Printf("Error updating clans: %s\n", err)
		handleError(w, types.RESTError{StatusCode: http.StatusInternalServerError, Error: "Could not set clans"})
	}
}
