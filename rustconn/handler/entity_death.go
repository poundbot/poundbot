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

type EntityDeath struct {
	ls  string
	ras storage.RaidAlertsStore
}

func NewEntityDeath(ls string, ras storage.RaidAlertsStore) func(w http.ResponseWriter, r *http.Request) {
	ed := EntityDeath{ls, ras}
	return ed.Handle
}

// Handle manages incoming Rust entity death notices and sends them
// to the RaidAlertsStore and RaidAlerts channel
func (e *EntityDeath) Handle(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	account := context.Get(r, "account").(types.Account)

	decoder := json.NewDecoder(r.Body)
	var ed types.EntityDeath
	err := decoder.Decode(&ed)
	if err != nil {
		log.Println(e.ls + err.Error())
		handleError(w, e.ls, types.RESTError{
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
