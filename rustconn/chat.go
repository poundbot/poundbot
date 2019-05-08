package rustconn

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/blang/semver"
	"github.com/poundbot/poundbot/pbclock"
	"github.com/poundbot/poundbot/types"
)

var iclock = pbclock.Clock

type chatQueue interface {
	GetGameServerMessage(sk string, to time.Duration) (types.ChatMessage, bool)
}

type discordChat struct {
	ClanTag     string
	DisplayName string
	Message     string
}

type deprecatedChat struct {
	types.ChatMessage
	SteamID uint64
}

func (d *deprecatedChat) upgrade() {
	if d.SteamID == 0 {
		return
	}
	d.PlayerID = fmt.Sprintf("%d", d.SteamID)
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
	in         chan types.ChatMessage
	timeout    time.Duration
	logger     *log.Logger
	minVersion semver.Version
}

// NewChat initializes a chat handler and returns it
//
// ls is the log symbol
// in is the channel for server -> discord
// out is the channel for discord -> server
func newChat(ls string, cq chatQueue, in chan types.ChatMessage) func(w http.ResponseWriter, r *http.Request) {

	c := chat{
		cqs:        cq,
		in:         in,
		timeout:    10 * time.Second,
		logger:     &log.Logger{},
		minVersion: semver.Version{Major: 1, Patch: 1},
	}

	c.logger.SetPrefix(ls)
	c.logger.SetOutput(os.Stdout)

	return c.Handle
}

// Handle manages Rust <-> discord chat requests and logging
// Discord -> Rust is through the "out" chan and Rust -> Discord is
// through the "in" chan.
//
// HTTP POST requests are sent to the "in" chan
//
// HTTP GET requests wait for messages and disconnect with http.StatusNoContent
// after timeout seconds.
func (c *chat) Handle(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	version, err := semver.Make(r.Header.Get("X-PoundBotBetterChat-Version"))
	if err == nil && version.LT(c.minVersion) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("PoundBotBetterChat must be updated. Please download the latest version at " + upgradeURL))
		return
	}

	sc, err := getServerContext(r.Context())
	if err != nil {
		c.logger.Printf("[%s](%s:%s) Can't find server: %s", sc.requestUUID, sc.account.ID.Hex(), sc.serverKey, err.Error())
		handleError(w, types.RESTError{
			Error:      "Error finding server identity",
			StatusCode: http.StatusInternalServerError,
		})
		return
	}

	switch r.Method {
	case http.MethodPost:
		decoder := json.NewDecoder(r.Body)
		var m deprecatedChat

		err := decoder.Decode(&m)
		if err != nil {
			c.logger.Printf("[%s](%s:%s) Invalid JSON: %s", sc.requestUUID, sc.account.ID.Hex(), sc.server.Name, err.Error())
			handleError(w, types.RESTError{
				Error:      "Invalid request",
				StatusCode: http.StatusBadRequest,
			})
			return
		}

		m.upgrade()
		m.PlayerID = fmt.Sprintf("%s:%s", sc.game, m.PlayerID)

		found, clan := sc.server.UsersClan([]string{m.PlayerID})
		if found {
			m.ClanTag = clan.Tag
		}

		for _, s := range sc.account.Servers {
			if s.Key == sc.serverKey {
				m.ChannelID = s.ChatChanID
				select {
				case c.in <- m.ChatMessage:
					return
				case <-time.After(c.timeout):
					return
				}
			}
		}

	case http.MethodGet:
		m, found := c.cqs.GetGameServerMessage(sc.serverKey, c.timeout)
		if !found {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		b, err := json.Marshal(newDiscordChat(m))
		if err != nil {
			c.logger.Printf("[%s] %s", sc.requestUUID, err.Error())
			return
		}

		w.Write(b)
	}
}
