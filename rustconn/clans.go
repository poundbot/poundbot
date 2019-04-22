package rustconn

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/poundbot/poundbot/storage"
	"github.com/poundbot/poundbot/types"
)

type serverClan struct {
	Tag        string
	ClanTag    string
	Owner      string
	OwnerID    string
	Members    []string
	Moderators []string
}

func (s serverClan) ToClan() types.Clan {
	c := types.Clan{}
	c.Members = s.Members
	c.Moderators = s.Moderators
	if s.Owner != "" {
		// RustIO Clan
		c.OwnerID = s.Owner
		c.Tag = s.Tag
		return c
	}

	c.OwnerID = s.OwnerID
	c.Tag = s.ClanTag
	return c
}

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

	var sClans []serverClan
	err := decoder.Decode(&sClans)
	if err != nil {
		log.Println(err.Error())
		handleError(w, types.RESTError{StatusCode: http.StatusBadRequest, Error: "Could not decode clans"})
		return
	}

	var clans = make([]types.Clan, len(sClans))

	for i := range sClans {
		clans[i] = sClans[i].ToClan()
		clans[i].SetGame(game)
	}

	err = c.as.SetClans(serverKey, clans)
	if err != nil {
		c.logger.Printf("Error updating clans: %s\n", err)
		handleError(w, types.RESTError{StatusCode: http.StatusInternalServerError, Error: "Could not set clans"})
	}
}

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
	game := r.Context().Value(contextKeyGame).(string)
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
		var sClan serverClan
		err := decoder.Decode(&sClan)
		if err != nil {
			c.logger.Println(err.Error())
			return
		}

		clan := sClan.ToClan()

		clan.SetGame(game)

		c.as.AddClan(serverKey, clan)
	}
}
