package gameapi

import (
	"encoding/json"
	"net/http"

	"github.com/poundbot/poundbot/types"
	"github.com/gorilla/mux"
)

type playerIDs []string

type registeredPlayers struct{}

func initPlayers(api *mux.Router) {
	rp := registeredPlayers{}
	api.HandleFunc("/players/registered", rp.handle).Methods(http.MethodGet)
}

func (p *registeredPlayers) handle(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	sc, err := getServerContext(r.Context())
	if err != nil {
		log.Printf("[%s](%s:%s) Can't find server: %s", sc.requestUUID, sc.account.ID.Hex(), sc.serverKey, err.Error())
		handleError(w, types.RESTError{
			Error:      "Error finding server identity",
			StatusCode: http.StatusInternalServerError,
		})
		return
	}

	switch r.Method {
	case http.MethodGet:
		b, err := json.Marshal(sc.account.GetRegisteredPlayerIDs(sc.game))
		if err != nil {
			log.Printf("[%s] %s", sc.requestUUID, err.Error())
			return
		}

		w.Write(b)

	}
}
