package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"bitbucket.org/mrpoundsign/poundbot/storage"
	"bitbucket.org/mrpoundsign/poundbot/types"
)

type discordAuth struct {
	das    storage.DiscordAuthsStore
	us     storage.UsersStore
	dac    chan types.DiscordAuth
	logger *log.Logger
}

func NewDiscordAuth(logPrefix string, das storage.DiscordAuthsStore, us storage.UsersStore, dac chan types.DiscordAuth) func(w http.ResponseWriter, r *http.Request) {
	da := discordAuth{das: das, us: us, dac: dac, logger: &log.Logger{}}
	da.logger.SetPrefix(logPrefix)
	return da.Handle
}

// Handle takes Discord verification requests from the Rust server
// and sends them to the DiscordAuthsStore and DiscordAuth channel
func (da *discordAuth) Handle(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	var t types.DiscordAuth
	err := decoder.Decode(&t)
	if err != nil {
		da.logger.Println(err.Error())
		return
	}

	// da.logger.Printf("User Auth Request: %v from %v\n", t, r.Body)

	var u types.User
	err = da.us.Get(t.SteamID, &u)
	if err == nil {
		handleError(w, types.RESTError{
			StatusCode: http.StatusMethodNotAllowed,
			Error:      fmt.Sprintf("%s is linked to you.", u.DiscordName),
		})
		return
	} else if t.DiscordName == "check" {
		handleError(w, types.RESTError{
			StatusCode: http.StatusNotFound,
			Error:      "Account is not linked to discord.",
		})
		return
	}

	err = da.das.Upsert(t)
	if err == nil {
		da.dac <- t
	} else {
		da.logger.Println(err.Error())
	}
}
