package rustconn

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/poundbot/poundbot/storage"
	"github.com/poundbot/poundbot/types"
)

type clans struct {
	as     storage.AccountsStore
	logger *log.Logger
}

func NewClans(logPrefix string, as storage.AccountsStore) func(w http.ResponseWriter, r *http.Request) {
	c := clans{as: as, logger: &log.Logger{}}
	c.logger.SetPrefix(logPrefix)
	c.logger.SetOutput(os.Stdout)
	return c.Handle
}

// Handle manages clans sync HTTP requests from the Rust server
// These requests are a complete refresh of all clans
func (c *clans) Handle(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	game := r.Context().Value(contextKeyGame).(string)
	account := r.Context().Value(contextKeyAccount).(types.Account)
	serverKey := r.Context().Value(contextKeyServerKey).(string)
	requestUUID := r.Context().Value(contextKeyRequestUUID).(string)
	log.Printf("[%s] %s: Updating all clans for %s\n", requestUUID, game, account.ID)

	decoder := json.NewDecoder(r.Body)

	var sClans []types.Clan
	err := decoder.Decode(&sClans)
	if err != nil {
		log.Println(err.Error())
		handleError(w, types.RESTError{StatusCode: http.StatusBadRequest, Error: "Could not decode clans"})
		return
	}

	for i := range sClans {
		sClans[i].SetGame(game)
	}

	err = c.as.SetClans(serverKey, sClans)
	if err != nil {
		c.logger.Printf("Error updating clans: %s\n", err)
		handleError(w, types.RESTError{StatusCode: http.StatusInternalServerError, Error: "Could not set clans"})
	}
}
