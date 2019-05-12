package rustconn

import (
	"encoding/json"
	"net/http"

	"github.com/poundbot/poundbot/types"
)

type playerIDs []string

type registeredPlayers struct{}

func newRegisteredPlayers() func(w http.ResponseWriter, r *http.Request) {
	rp := registeredPlayers{}
	return rp.Handle
}

func (p *registeredPlayers) Handle(w http.ResponseWriter, r *http.Request) {
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
