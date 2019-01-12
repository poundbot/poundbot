package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"bitbucket.org/mrpoundsign/poundbot/storage"
	"bitbucket.org/mrpoundsign/poundbot/types"
	"github.com/gorilla/context"
)

type Clans struct {
	ls string
	as storage.AccountsStore
}

func NewClans(ls string, as storage.AccountsStore) func(w http.ResponseWriter, r *http.Request) {
	clans := Clans{ls, as}
	return clans.Handle
}

// Handle manages clans sync HTTP requests from the Rust server
// These requests are a complete refresh of all clans
func (c *Clans) Handle(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	serverKey := context.Get(r, "serverKey").(string)
	log.Printf(c.ls+"clansHandler: Updating all clans for %s\n", serverKey)

	decoder := json.NewDecoder(r.Body)
	var t []types.ServerClan
	err := decoder.Decode(&t)
	if err != nil {
		log.Println(c.ls + err.Error())
		handleError(w, c.ls, types.RESTError{StatusCode: http.StatusBadRequest, Error: "Could not decode clans"})
		return
	}

	clanCount := len(t)
	clans := make([]types.Clan, clanCount)
	for i, sc := range t {
		cl, err := types.ClanFromServerClan(sc)
		if err != nil {
			log.Printf(c.ls+"clansHandler Error: %v\n", err)
			handleError(w, c.ls, types.RESTError{
				StatusCode: http.StatusBadRequest,
				Error:      "Error processing clan data",
			})
			return
		}
		clans[i] = *cl
	}

	err = c.as.SetClans(serverKey, clans)
	if err != nil {
		fmt.Printf("Error updating clans: %s\n", err)
		handleError(w, c.ls, types.RESTError{StatusCode: http.StatusInternalServerError, Error: "Could not set clans"})
	}
}
