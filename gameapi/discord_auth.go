package gameapi

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/poundbot/poundbot/types"
)

type discordAuthenticator interface {
	AuthDiscord(types.DiscordAuth)
}

type daAuthUpserter interface {
	Upsert(types.DiscordAuth) error
}

type daUserGetter interface {
	GetByPlayerID(string) (types.User, error)
}

type discordAuth struct {
	dau daAuthUpserter
	us  daUserGetter
	da  discordAuthenticator
}

type deprecatedDiscordAuth struct {
	types.DiscordAuth
	SteamID uint64
}

func (d *deprecatedDiscordAuth) upgrade() {
	if d.SteamID == 0 {
		return
	}
	d.PlayerID = fmt.Sprintf("%d", d.SteamID)
}

func initDiscordAuth(dau daAuthUpserter, us daUserGetter, dah discordAuthenticator, api *mux.Router) {
	da := discordAuth{dau: dau, us: us, da: dah}
	api.HandleFunc("/discord_auth", da.handle)
}

// handle takes Discord verification requests from the Rust server
// and sends them to the DiscordAuthsStore and DiscordAuth channel
func (da *discordAuth) handle(w http.ResponseWriter, r *http.Request) {
	game := r.Context().Value(contextKeyGame).(string)
	account := r.Context().Value(contextKeyAccount).(types.Account)

	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	var dAuth deprecatedDiscordAuth

	err := decoder.Decode(&dAuth)
	if err != nil {
		log.Println(err.Error())
		return
	}

	dAuth.upgrade()
	dAuth.PlayerID = fmt.Sprintf("%s:%s", game, dAuth.PlayerID)

	user, err := da.us.GetByPlayerID(dAuth.PlayerID)
	if err == nil {
		handleError(w, types.RESTError{
			StatusCode: http.StatusMethodNotAllowed,
			Error:      fmt.Sprintf("%s is linked to you.", user.DiscordName),
		})
		return
	}

	if dAuth.DiscordName == "check" {
		handleError(w, types.RESTError{
			StatusCode: http.StatusNotFound,
			Error:      "Account is not linked to discord.",
		})
		return
	}

	dAuth.GuildSnowflake = account.GuildSnowflake

	err = da.dau.Upsert(dAuth.DiscordAuth)
	if err != nil {
		log.Println(err.Error())
		return
	}
	da.da.AuthDiscord(dAuth.DiscordAuth)
}
