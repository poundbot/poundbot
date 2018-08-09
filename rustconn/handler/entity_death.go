package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"bitbucket.org/mrpoundsign/poundbot/db"
	"bitbucket.org/mrpoundsign/poundbot/types"
)

type EntityDeath struct {
	ls  string
	ras db.RaidAlertsStore
}

func NewEntityDeath(ls string, ras db.RaidAlertsStore) func(w http.ResponseWriter, r *http.Request) {
	ed := EntityDeath{ls, ras}
	return ed.Handle
}

// Handle manages incoming Rust entity death notices and sends them
// to the RaidAlertsStore and RaidAlerts channel
func (e *EntityDeath) Handle(w http.ResponseWriter, r *http.Request) {
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
	e.ras.AddInfo(ed)
}
