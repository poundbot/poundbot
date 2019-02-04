package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"bitbucket.org/mrpoundsign/poundbot/chatcache"
	ptime "bitbucket.org/mrpoundsign/poundbot/time"
	"bitbucket.org/mrpoundsign/poundbot/types"
)

// A Chat is for handling discord <-> rust chat
type chat struct {
	ccache *chatcache.ChatCache
	in     chan types.ChatMessage
	sleep  time.Duration
	logger *log.Logger
}

// NewChat initializes a chat handler and returns it
//
// ls is the log symbol
// in is the channel for server -> discord
// out is the channel for discord -> server
func NewChat(ls string, ccache *chatcache.ChatCache, in chan types.ChatMessage) func(w http.ResponseWriter, r *http.Request) {
	c := chat{ccache: ccache, in: in, sleep: 1 * time.Minute, logger: &log.Logger{}}

	c.logger.SetPrefix(ls)

	return c.Handle
}

// Handle manages Rust <-> discord chat requests and logging
// Discord -> Rust is through the "out" chan and Rust -> Discord is
// through the "in" chan.
//
// HTTP POST requests are sent to the "in" chan
//
// HTTP GET requests wait for messages and disconnect with http.StatusNoContent
// after sleep seconds.
func (c *chat) Handle(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	serverKey := r.Context().Value("serverKey").(string)
	requestUUID := r.Context().Value("requestUUID").(string)
	account := r.Context().Value("account").(types.Account)

	switch r.Method {
	case http.MethodPost:
		decoder := json.NewDecoder(r.Body)
		var m types.ChatMessage
		err := decoder.Decode(&m)
		if err != nil {
			c.logger.Printf("[%s] Invalid JSON: %s", requestUUID, err.Error())
			return
		}

		m.Source = types.ChatSourceRust

		if m.CreatedAt.Equal(time.Time{}) {
			m.CreatedAt = ptime.Clock().Now().UTC()
		}

		for _, s := range account.Servers {
			if s.Key == serverKey {
				m.ChannelID = s.ChatChanID
				select {
				case c.in <- m:
					return
				case <-time.After(c.sleep):
					return
				}
			}
		}

	case http.MethodGet:
		ch := c.ccache.GetOutChannel(serverKey)
		select {
		case m := <-ch:
			b, err := json.Marshal(m)
			if err != nil {
				c.logger.Printf("[%s] %s", requestUUID, err.Error())
				return
			}

			w.Write(b)
		case <-time.After(c.sleep):
			return
		}

	default:
		methodNotAllowed(w)
	}
}
