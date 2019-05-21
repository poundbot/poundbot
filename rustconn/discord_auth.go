package rustconn

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/poundbot/poundbot/types"
)

type daAuthUpserter interface {
	Upsert(types.DiscordAuth) error
}

type daUserGetter interface {
	GetByPlayerID(string) (types.User, error)
}

type discordAuth struct {
	dau daAuthUpserter
	us  daUserGetter
	dac chan<- types.DiscordAuth
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

func newDiscordAuth(dau daAuthUpserter, us daUserGetter, dac chan<- types.DiscordAuth) func(w http.ResponseWriter, r *http.Request) {
	da := discordAuth{dau: dau, us: us, dac: dac}
	return da.Handle
}

// Handle takes Discord verification requests from the Rust server
// and sends them to the DiscordAuthsStore and DiscordAuth channel
func (da *discordAuth) Handle(w http.ResponseWriter, r *http.Request) {
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
	if err == nil {
		da.dac <- dAuth.DiscordAuth
	} else {
		log.Println(err.Error())
	}
}
