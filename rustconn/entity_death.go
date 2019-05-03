package rustconn

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
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
	logger     *log.Logger
	minVersion semver.Version
}

func NewEntityDeath(logPrefix string, ras storage.RaidAlertsStore) func(w http.ResponseWriter, r *http.Request) {
	ed := entityDeath{ras: ras, logger: &log.Logger{}, minVersion: semver.Version{Major: 1}}
	ed.logger.SetPrefix(logPrefix)
	ed.logger.SetOutput(os.Stdout)
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

	game := r.Context().Value(contextKeyGame).(string)
	serverKey := r.Context().Value(contextKeyServerKey).(string)
	requestUUID := r.Context().Value(contextKeyRequestUUID).(string)
	account := r.Context().Value(contextKeyAccount).(types.Account)
	server, err := account.ServerFromKey(serverKey)
	if err != nil {
		e.logger.Printf("[%s](%s) Invalid JSON: %s", requestUUID, account.ID.Hex(), err.Error())
		handleError(w, types.RESTError{
			Error:      "Error processing request: Could not find server from key.",
			StatusCode: http.StatusInternalServerError,
		})
		return
	}

	decoder := json.NewDecoder(r.Body)
	var ed deprecatedEntityDeath
	err = decoder.Decode(&ed)
	if err != nil {
		e.logger.Printf("[%s](%s:%s) Invalid JSON: %s", requestUUID, account.ID.Hex(), server.Name, err.Error())
		handleError(w, types.RESTError{
			Error:      "Invalid request",
			StatusCode: http.StatusBadRequest,
		})
		return
	}

	ed.upgrade()

	for i := range ed.OwnerIDs {
		ed.OwnerIDs[i] = fmt.Sprintf("%s:%s", game, ed.OwnerIDs[i])
	}

	if ed.ServerName == "" {
		ed.ServerName = server.Name
	}
	alertAt := 10 * time.Second
	if len(account.Servers) != 0 {
		sAlertAt, err := time.ParseDuration(server.RaidDelay)
		if err == nil {
			alertAt = sAlertAt
		}
	}
	e.ras.AddInfo(alertAt, ed.EntityDeath)
}
