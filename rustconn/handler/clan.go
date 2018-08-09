package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"bitbucket.org/mrpoundsign/poundbot/db"
	"bitbucket.org/mrpoundsign/poundbot/types"
	"github.com/gorilla/mux"
)

type Clan struct {
	ls string
	cs db.ClansStore
	us db.UsersStore
}

func NewClan(ls string, cs db.ClansStore, us db.UsersStore) func(w http.ResponseWriter, r *http.Request) {
	clan := Clan{ls, cs, us}
	return clan.Handle
}

// Handle manages individual clan REST requests form the Rust server
func (c *Clan) Handle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tag := vars["tag"]

	switch r.Method {
	case http.MethodDelete:
		log.Printf(c.ls+"clanHandler: Removing clan %s\n", tag)
		c.cs.Remove(tag)
		c.us.RemoveClan(tag)
		return
	case http.MethodPut:
		log.Printf(c.ls+"clanHandler: Updating clan %s\n", tag)
		decoder := json.NewDecoder(r.Body)
		var t types.ServerClan
		err := decoder.Decode(&t)
		if err != nil {
			log.Println(c.ls + err.Error())
			return
		}

		clan, err := types.ClanFromServerClan(t)
		if err != nil {
			handleError(w, c.ls, types.RESTError{
				StatusCode: http.StatusBadRequest,
				Error:      "Error processing clan data",
			})
			return
		}

		c.cs.Upsert(*clan)
	}
}
