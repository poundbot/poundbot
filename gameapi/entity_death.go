package gameapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/blang/semver"
	"github.com/gorilla/mux"
	"github.com/poundbot/poundbot/types"
)

type raidAlertAdder interface {
	AddInfo(alertIn, validUntil time.Duration, ed types.EntityDeath) error
}

type deprecatedEntityDeath struct {
	types.EntityDeath
	Owners []int64
}

func (d *deprecatedEntityDeath) upgrade() {
	if len(d.Owners) == 0 {
		return
	}

	d.OwnerIDs = make([]string, len(d.Owners))

	for i := range d.Owners {
		d.OwnerIDs[i] = fmt.Sprintf("%d", d.Owners[i])
	}
}

type entityDeath struct {
	raa        raidAlertAdder
	minVersion semver.Version
}

func initEntityDeath(api *mux.Router, path string, raa raidAlertAdder) {
	ed := entityDeath{raa: raa, minVersion: semver.Version{Major: 1}}
	api.HandleFunc(path, ed.handle)
}

// handle manages incoming Rust entity death notices and sends them
// to the RaidAlertsStore and RaidAlerts channel
func (e *entityDeath) handle(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	sc, err := getServerContext(r.Context())

	edLog := logWithRequest(r.RequestURI, sc)

	if err != nil {
		edLog.WithError(err).Info("Can't find server")
		handleError(w, types.RESTError{
			Error:      "Error finding server identity",
			StatusCode: http.StatusInternalServerError,
		})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var ed deprecatedEntityDeath
	err = decoder.Decode(&ed)
	if err != nil {
		edLog.WithError(err).Error("Invalid JSON")
		handleError(w, types.RESTError{
			Error:      "Invalid request",
			StatusCode: http.StatusBadRequest,
		})
		return
	}

	ed.upgrade()

	for i := range ed.OwnerIDs {
		ed.OwnerIDs[i] = fmt.Sprintf("%s:%s", sc.game, ed.OwnerIDs[i])
	}

	if len(ed.ServerName) == 0 {
		ed.ServerName = sc.server.Name
	}

	if len(ed.ServerKey) == 8 {
		ed.ServerKey = sc.server.Key
	}

	alertAt := 10 * time.Second
	validUntil := 15 * time.Minute

	sAlertAt, err := time.ParseDuration(sc.server.RaidDelay)
	if err == nil {
		alertAt = sAlertAt
	}

	sValidUntil, err := time.ParseDuration(sc.server.RaidNotifyFrequency)
	if err == nil {
		validUntil = sValidUntil
	}

	e.raa.AddInfo(alertAt, validUntil, ed.EntityDeath)
}
