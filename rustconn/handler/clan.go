package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"bitbucket.org/mrpoundsign/poundbot/storage"
	"bitbucket.org/mrpoundsign/poundbot/types"
	"github.com/gorilla/mux"
)

type clan struct {
	as     storage.AccountsStore
	logger *log.Logger
}

func NewClan(logPrefix string, as storage.AccountsStore, us storage.UsersStore) func(w http.ResponseWriter, r *http.Request) {
	c := clan{as: as, logger: &log.Logger{}}
	c.logger.SetPrefix(logPrefix)
	return c.Handle
}

// Handle manages individual clan REST requests form the Rust server
func (c *clan) Handle(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	// account := context.Get(r, "account").(types.Account)
	serverKey := r.Context().Value("serverKey").(string)

	vars := mux.Vars(r)
	tag := vars["tag"]

	switch r.Method {
	case http.MethodDelete:
		c.logger.Printf("clanHandler: Removing clan %s for %s\n", tag, serverKey)
		c.as.RemoveClan(serverKey, tag)
		return
	case http.MethodPut:
		c.logger.Printf("clanHandler: Updating clan %s for %s\n", tag, serverKey)
		decoder := json.NewDecoder(r.Body)
		var t types.ServerClan
		err := decoder.Decode(&t)
		if err != nil {
			c.logger.Println(err.Error())
			return
		}

		clan, err := types.ClanFromServerClan(t)
		if err != nil {
			handleError(w, types.RESTError{
				StatusCode: http.StatusBadRequest,
				Error:      "Error processing clan data",
			})
			return
		}

		c.as.AddClan(serverKey, *clan)
	}
}
