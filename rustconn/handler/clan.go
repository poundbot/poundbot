package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"bitbucket.org/mrpoundsign/poundbot/storage"
	"bitbucket.org/mrpoundsign/poundbot/types"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
)

type Clan struct {
	ls string
	as storage.AccountsStore
}

func NewClan(ls string, as storage.AccountsStore, us storage.UsersStore) func(w http.ResponseWriter, r *http.Request) {
	clan := Clan{ls, as}
	return clan.Handle
}

// Handle manages individual clan REST requests form the Rust server
func (c *Clan) Handle(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	defer context.Clear(r)
	// account := context.Get(r, "account").(types.Account)
	serverKey := context.Get(r, "serverKey").(string)
	log.Printf("Server key is %s\n", serverKey)

	vars := mux.Vars(r)
	tag := vars["tag"]

	switch r.Method {
	case http.MethodDelete:
		log.Printf(c.ls+"clanHandler: Removing clan %s\n", tag)
		c.as.RemoveClan(serverKey, tag)
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

		c.as.AddClan(serverKey, *clan)
	}
}
