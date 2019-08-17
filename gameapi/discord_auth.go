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

type discordAuthRequest struct {
	types.DiscordAuth
}

func initDiscordAuth(dau daAuthUpserter, us daUserGetter, dah discordAuthenticator, api *mux.Router) {
	da := discordAuth{dau: dau, us: us, da: dah}
	api.HandleFunc("/discord_auth", da.handle).Methods("PUT")
	api.HandleFunc("/discord_auth/check/{player_id}", da.checkPlayer).Methods("GET")
}

// handle takes Discord verification requests from the Rust server
// and sends them to the DiscordAuthsStore and DiscordAuth channel
func (da *discordAuth) handle(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	game := r.Context().Value(contextKeyGame).(string)
	account := r.Context().Value(contextKeyAccount).(types.Account)

	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	var dAuth discordAuthRequest

	err := decoder.Decode(&dAuth)
	if err != nil {
		log.Println(err.Error())
		return
	}

	dAuth.PlayerID = fmt.Sprintf("%s:%s", game, dAuth.PlayerID)

	user, err := da.us.GetByPlayerID(dAuth.PlayerID)
	if err == nil {
		handleError(w, types.RESTError{
			StatusCode: http.StatusConflict,
			Error:      fmt.Sprintf("%s is already registered.", user.DiscordName),
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

func (da *discordAuth) checkPlayer(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	sc, err := getServerContext(r.Context())

	params := mux.Vars(r)
	cpLog := logWithRequest(sc).WithField("pID", params["player_id"])

	if err != nil {
		cpLog.Info("Can't find server")
		handleError(w, types.RESTError{
			Error:      "Error finding server identity",
			StatusCode: http.StatusInternalServerError,
		})
		return
	}

	cpLog.Trace("Checking player")
	_, err = da.us.GetByPlayerID(fmt.Sprintf("%s:%s", sc.game, params["player_id"]))
	if err != nil {
		cpLog.Trace("Player not found")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	cpLog.Trace("Player found")
}
