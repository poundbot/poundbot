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

func newClans(logPrefix string, as storage.AccountsStore) func(w http.ResponseWriter, r *http.Request) {
	c := clans{as: as, logger: &log.Logger{}}
	c.logger.SetPrefix(logPrefix)
	c.logger.SetOutput(os.Stdout)
	return c.Handle
}

// Handle manages clans sync HTTP requests from the Rust server
// These requests are a complete refresh of all clans
func (c *clans) Handle(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	sc, err := getServerContext(r.Context())
	if err != nil {
		c.logger.Printf("[%s](%s:%s) Can't find server: %s", sc.requestUUID, sc.account.ID.Hex(), sc.serverKey, err.Error())
		handleError(w, types.RESTError{
			Error:      "Error finding server identity",
			StatusCode: http.StatusInternalServerError,
		})
		return
	}
	log.Printf("[%s] %s: Updating all clans for %s:%s\n", sc.requestUUID, sc.game, sc.account.ID.Hex(), sc.server.Name)

	decoder := json.NewDecoder(r.Body)

	var sClans []serverClan
	err = decoder.Decode(&sClans)
	if err != nil {
		log.Println(err.Error())
		handleError(w, types.RESTError{StatusCode: http.StatusBadRequest, Error: "Could not decode clans"})
		return
	}

	var clans = make([]types.Clan, len(sClans))

	for i := range sClans {
		clans[i] = sClans[i].ToClan()
		clans[i].SetGame(sc.game)
	}

	err = c.as.SetClans(sc.serverKey, clans)
	if err != nil {
		c.logger.Printf("Error updating clans: %s\n", err)
		handleError(w, types.RESTError{StatusCode: http.StatusInternalServerError, Error: "Could not set clans"})
	}
}

type clan struct {
	as     storage.AccountsStore
	logger *log.Logger
}

func newClan(logPrefix string, as storage.AccountsStore, us storage.UsersStore) func(w http.ResponseWriter, r *http.Request) {
	c := clan{as: as, logger: &log.Logger{}}
	c.logger.SetPrefix(logPrefix)
	c.logger.SetOutput(os.Stdout)
	return c.Handle
}

// Handle manages individual clan REST requests form the Rust server
func (c *clan) Handle(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	sc, err := getServerContext(r.Context())
	if err != nil {
		c.logger.Printf("[%s](%s:%s) Can't find server: %s", sc.requestUUID, sc.account.ID.Hex(), sc.serverKey, err.Error())
		handleError(w, types.RESTError{
			Error:      "Error finding server identity",
			StatusCode: http.StatusInternalServerError,
		})
		return
	}

	vars := mux.Vars(r)
	tag := vars["tag"]

	switch r.Method {
	case http.MethodDelete:
		log.Printf("[%s] %s: Removing clan \"%s\" for %s:%s\n", sc.requestUUID, sc.game, tag, sc.account.ID.Hex(), sc.server.Name)
		err := c.as.RemoveClan(sc.serverKey, tag)
		if err != nil {
			handleError(w, types.RESTError{
				Error:      "Could not remove clan",
				StatusCode: http.StatusInternalServerError,
			})
			log.Printf("[%s] %s: Error removing clan \"%s\" for %s:%s: %v\n", sc.requestUUID, sc.game, tag, sc.account.ID.Hex(), sc.server.Name, err)
		}
		return
	case http.MethodPut:
		log.Printf("[%s] %s: Updating clan \"%s\" for %s:%s\n", sc.requestUUID, sc.game, tag, sc.account.ID.Hex(), sc.server.Name)
		decoder := json.NewDecoder(r.Body)
		var sClan serverClan
		err := decoder.Decode(&sClan)
		if err != nil {
			log.Printf("[%s] %s: Error decoding clan \"%s\" for %s:%s: %v\n", sc.requestUUID, sc.game, tag, sc.account.ID.Hex(), sc.server.Name, err)
			handleError(w, types.RESTError{
				Error:      "Could not decode clan data",
				StatusCode: http.StatusBadRequest,
			})
			return
		}

		clan := sClan.ToClan()

		clan.SetGame(sc.game)

		err = c.as.AddClan(sc.serverKey, clan)
		if err != nil {
			handleError(w, types.RESTError{
				Error:      "Could not add clan",
				StatusCode: http.StatusInternalServerError,
			})
			log.Printf("[%s] %s: Error adding clan \"%s\" for %s:%s: %v\n", sc.requestUUID, sc.game, tag, sc.account.ID.Hex(), sc.server.Name, err)
		}
	}
}
