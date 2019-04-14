package rustconn

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/poundbot/poundbot/storage"
	"github.com/poundbot/poundbot/types"
	"github.com/gorilla/mux"
)

type clan struct {
	as     storage.AccountsStore
	logger *log.Logger
}

func NewClan(logPrefix string, as storage.AccountsStore, us storage.UsersStore) func(w http.ResponseWriter, r *http.Request) {
	c := clan{as: as, logger: &log.Logger{}}
	c.logger.SetPrefix(logPrefix)
	c.logger.SetOutput(os.Stdout)
	return c.Handle
}

// Handle manages individual clan REST requests form the Rust server
func (c *clan) Handle(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	account := r.Context().Value(contextKeyAccount).(types.Account)
	serverKey := r.Context().Value(contextKeyServerKey).(string)
	requestUUID := r.Context().Value(contextKeyRequestUUID).(string)
	server, err := account.ServerFromKey(serverKey)
	if err != nil {
		handleError(w, types.RESTError{
			StatusCode: http.StatusInternalServerError,
			Error:      "Internal error",
		})
	}

	vars := mux.Vars(r)
	tag := vars["tag"]

	switch r.Method {
	case http.MethodDelete:
		c.logger.Printf("[%s] Removing clan %s for %s:%s\n", requestUUID, tag, account.ID, server.Name)
		c.as.RemoveClan(serverKey, tag)
		return
	case http.MethodPut:
		c.logger.Printf("[%s] Updating clan %s for %s:%s\n", requestUUID, tag, account.ID, server.Name)
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
