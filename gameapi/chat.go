package gameapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/blang/semver"
	"github.com/gorilla/mux"
	"github.com/poundbot/poundbot/pbclock"
	"github.com/poundbot/poundbot/types"
)

var iclock = pbclock.Clock

type discordMessager interface {
	SendChatMessage(types.ChatMessage)
}

type chatQueue interface {
	GetGameServerMessage(sk, tag string, to time.Duration) (types.ChatMessage, bool)
}

type discordChat struct {
	ClanTag     string
	DisplayName string
	Message     string
}

func newDiscordChat(cm types.ChatMessage) discordChat {
	return discordChat{
		ClanTag:     cm.ClanTag,
		DisplayName: cm.DisplayName,
		Message:     cm.Message,
	}
}

// A Chat is for handling discord <-> rust chat
type chat struct {
	cqs        chatQueue
	dm         discordMessager
	timeout    time.Duration
	minVersion semver.Version
}

// initChat initializes a chat handler and returns it
//
// cq is the chatQueue for reading messages from
// in is the channel for server -> discord
func initChat(cq chatQueue, dm discordMessager, api *mux.Router) {
	c := chat{
		cqs:        cq,
		dm:         dm,
		timeout:    10 * time.Second,
		minVersion: semver.Version{Major: 1, Patch: 3},
	}

	api.HandleFunc("/chat", c.handle).Methods(http.MethodGet, http.MethodPost)
}

// handle manages Rust <-> discord chat requests and logging
//
// HTTP POST requests are sent to the "in" chan
//
// HTTP GET requests wait for messages and disconnect with http.StatusNoContent
// after timeout seconds.
func (c *chat) handle(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	version, err := semver.Make(r.Header.Get("X-PoundBotChatRelay-Version"))
	if err == nil && version.LT(c.minVersion) {
		w.WriteHeader(http.StatusBadRequest)
		if _, err := w.Write([]byte("PoundBotChatRelay must be updated. Please download the latest version at " + upgradeURL)); err != nil {
			log.WithError(err).Error("Could not write output")
		}
		return
	}

	sc, err := getServerContext(r.Context())
	if err != nil {
		log.Info(fmt.Sprintf("[%s](%s:%s) Can't find server: %s", sc.requestUUID, sc.account.ID.Hex(), sc.serverKey, err.Error()))
		handleError(w, types.RESTError{
			Error:      "Error finding server identity",
			StatusCode: http.StatusForbidden,
		})
		return
	}

	switch r.Method {
	case http.MethodGet:
		m, found := c.cqs.GetGameServerMessage(sc.serverKey, "chat", c.timeout)
		if !found {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		b, err := json.Marshal(newDiscordChat(m))
		if err != nil {
			log.Printf("[%s] %s", sc.requestUUID, err.Error())
			return
		}

		w.Write(b)
	}
}
