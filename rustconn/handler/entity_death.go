package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"bitbucket.org/mrpoundsign/poundbot/storage"
	"bitbucket.org/mrpoundsign/poundbot/types"
	"github.com/gorilla/context"
)

type entityDeath struct {
	ras    storage.RaidAlertsStore
	logger log.Logger
}

func NewEntityDeath(logPrefix string, ras storage.RaidAlertsStore) func(w http.ResponseWriter, r *http.Request) {
	ed := entityDeath{ras: ras, logger: log.Logger{}}
	ed.logger.SetPrefix(logPrefix)
	return ed.Handle
}

// Handle manages incoming Rust entity death notices and sends them
// to the RaidAlertsStore and RaidAlerts channel
func (e *entityDeath) Handle(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	requestUUID := context.Get(r, "requestUUID").(string)
	account := context.Get(r, "account").(types.Account)

	decoder := json.NewDecoder(r.Body)
	var ed types.EntityDeath
	err := decoder.Decode(&ed)
	if err != nil {
		e.logger.Printf("[%s] Invalid JSON: %s", requestUUID, err.Error())
		handleError(w, types.RESTError{
			Error:      "Invalid request",
			StatusCode: http.StatusBadRequest,
		})
		return
	}
	alertAt := 10 * time.Second
	if len(account.Servers) != 0 {
		sAlertAt, err := time.ParseDuration(account.Servers[0].RaidDelay)
		if err == nil {
			alertAt = sAlertAt
		}
	}
	e.ras.AddInfo(alertAt, ed)
}
