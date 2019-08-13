package gameapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
	"github.com/poundbot/poundbot/types"
)

type discordRoleSetter interface {
	SetRole(types.RoleSet, time.Duration) error
}

type roles struct {
	drs     discordRoleSetter
	timeout time.Duration
}

func initRoles(drs discordRoleSetter, api *mux.Router) {
	r := roles{drs: drs, timeout: 10 * time.Second}

	api.HandleFunc("/roles/{role_name}", r.roleHandler).
		Methods(http.MethodPut)
}

func (rs roles) roleHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	role := vars["role_name"]

	sc, err := getServerContext(r.Context())
	if err != nil {
		handleError(w, types.RESTError{
			Error:      "Error finding server identity",
			StatusCode: http.StatusInternalServerError,
		})
		return
	}

	rhLog := log.WithFields(logrus.Fields{"uri": r.RequestURI, "requestID": sc.requestUUID, "accountID": sc.account.ID.Hex(), "serverName": sc.server.Name})

	decoder := json.NewDecoder(r.Body)
	var roleSet types.RoleSet

	err = decoder.Decode(&roleSet)
	if err != nil {
		rhLog.WithError(err).Error("Invalid JSON")
		if err := handleError(w, types.RESTError{
			Error:      "Invalid request",
			StatusCode: http.StatusBadRequest,
		}); err != nil {
			rhLog.WithError(err).Error("http response failed to write")
		}
		return
	}

	roleSet.Role = role
	roleSet.GuildID = sc.account.GuildSnowflake
	roleSet.SetGame(sc.game)

	if err := rs.drs.SetRole(roleSet, rs.timeout); err != nil {
		rhLog.Error("timed out sending message to channel")
		if err := handleError(w, types.RESTError{
			Error:      "internal error sending message to discord handler",
			StatusCode: http.StatusInternalServerError,
		}); err != nil {
			rhLog.WithError(err).Error("http response failed to write")
		}
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
