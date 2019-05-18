package rustconn

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/blang/semver"
	"github.com/poundbot/poundbot/storage"
	"github.com/poundbot/poundbot/types"
)

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
	ras        storage.RaidAlertsStore
	minVersion semver.Version
}

func newEntityDeath(ras storage.RaidAlertsStore) func(w http.ResponseWriter, r *http.Request) {
	ed := entityDeath{ras: ras, minVersion: semver.Version{Major: 1}}
	return ed.Handle
}

// Handle manages incoming Rust entity death notices and sends them
// to the RaidAlertsStore and RaidAlerts channel
func (e *entityDeath) Handle(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	version, err := semver.Make(r.Header.Get("X-PoundBotRaidAlerts-Version"))
	if err == nil && version.LT(e.minVersion) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("PoundBotRaidAlerts must be updated. Please download the latest version at " + upgradeURL))
		return
	}

	sc, err := getServerContext(r.Context())
	if err != nil {
		log.Printf("[%s](%s:%s) Can't find server: %s", sc.requestUUID, sc.account.ID.Hex(), sc.serverKey, err.Error())
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
		log.Printf("[%s](%s:%s) Invalid JSON: %s", sc.requestUUID, sc.account.ID.Hex(), sc.server.Name, err.Error())
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

	if ed.ServerName == "" {
		ed.ServerName = sc.server.Name
	}
	alertAt := 10 * time.Second

	sAlertAt, err := time.ParseDuration(sc.server.RaidDelay)
	if err == nil {
		alertAt = sAlertAt
	}

	e.ras.AddInfo(alertAt, ed.EntityDeath)
}
