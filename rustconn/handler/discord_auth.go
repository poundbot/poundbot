package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"bitbucket.org/mrpoundsign/poundbot/db"
	"bitbucket.org/mrpoundsign/poundbot/types"
)

type DiscordAuth struct {
	ls  string
	das db.DiscordAuthsStore
	us  db.UsersStore
	dac chan types.DiscordAuth
}

func NewDiscordAuth(ls string, das db.DiscordAuthsStore, us db.UsersStore, dac chan types.DiscordAuth) func(w http.ResponseWriter, r *http.Request) {
	da := DiscordAuth{ls: ls, das: das, us: us, dac: dac}
	return da.Handle
}

// Handle takes Discord verification requests from the Rust server
// and sends them to the DiscordAuthsStore and DiscordAuth channel
func (da *DiscordAuth) Handle(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var t types.DiscordAuth
	err := decoder.Decode(&t)
	if err != nil {
		log.Println(da.ls + err.Error())
		return
	}

	log.Printf(da.ls+"User Auth Request: %v from %v\n", t, r.Body)

	u, err := da.us.Get(t.SteamInfo)
	if err == nil {
		handleError(w, da.ls, types.RESTError{
			StatusCode: http.StatusMethodNotAllowed,
			Error:      fmt.Sprintf("%s is linked to you.", u.DiscordID),
		})
		return
	} else if t.DiscordID == "check" {
		handleError(w, da.ls, types.RESTError{
			StatusCode: http.StatusNotFound,
			Error:      "Account is not linked to discord.",
		})
		return
	}

	err = da.das.Upsert(t)
	if err == nil {
		da.dac <- t
	} else {
		log.Println(da.ls + err.Error())
	}
}