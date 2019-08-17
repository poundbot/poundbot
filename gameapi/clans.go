package gameapi

import (
	"encoding/json"
	"fmt"
	"net/http"

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
	as storage.AccountsStore
	us storage.UsersStore
}

func initClans(api *mux.Router, path string, as storage.AccountsStore, us storage.UsersStore) {
	c := clans{as: as, us: us}

	api.HandleFunc(path, c.rootHandler).
		Methods(http.MethodPut)

	api.HandleFunc(fmt.Sprintf("%s/{tag}", path), c.clanHandler).
		Methods(http.MethodDelete, http.MethodPut)
}

// Handle manages clans sync HTTP requests from the Rust server
// These requests are a complete refresh of all clans
func (c *clans) rootHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	sc, err := getServerContext(r.Context())
	rhLog := logWithRequest(r.RequestURI, sc)

	if err != nil {
		rhLog.WithError(err).Info("Can't find server")
		handleError(w, types.RESTError{
			Error:      "Error finding server identity",
			StatusCode: http.StatusInternalServerError,
		})
		return
	}

	rhLog.Info("Updating all clans")

	decoder := json.NewDecoder(r.Body)

	var sClans []serverClan
	err = decoder.Decode(&sClans)
	if err != nil {
		rhLog.WithError(err).Warn("Could not decode clans")
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
		rhLog.WithError(err).Error("Error updating clans")
		handleError(w, types.RESTError{StatusCode: http.StatusInternalServerError, Error: "Could not set clans"})
	}
}

// Handle manages individual clan REST requests form the Rust server
func (c *clans) clanHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	sc, err := getServerContext(r.Context())
	chLog := logWithRequest(r.RequestURI, sc)

	if err != nil {
		chLog.WithError(err).Info("Can't find server")
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
		chLog.Infof("Removing clan \"%s\"", tag)
		err := c.as.RemoveClan(sc.serverKey, tag)
		if err != nil {
			handleError(w, types.RESTError{
				Error:      "Could not remove clan",
				StatusCode: http.StatusInternalServerError,
			})
			chLog.WithError(err).Errorf("Error removing clan \"%s\"", tag)
		}
		return
	case http.MethodPut:
		chLog.Infof("Updating clan \"%s\"", tag)
		decoder := json.NewDecoder(r.Body)
		var sClan serverClan
		err := decoder.Decode(&sClan)
		if err != nil {
			chLog.WithError(err).Errorf("Error decoding clan \"%s\"", tag)
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
			chLog.WithError(err).Errorf("Error adding clan \"%s\"", tag)
		}
	}
}
