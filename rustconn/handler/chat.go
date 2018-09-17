package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"bitbucket.org/mrpoundsign/poundbot/storage"
	"bitbucket.org/mrpoundsign/poundbot/types"
	"github.com/gorilla/context"
)

// A Chat is for handling discord <-> rust chat
type Chat struct {
	d  bool
	ls string
	cs storage.ChatsStore
	in chan types.ChatMessage
	// out   chan types.ChatMessage
	sleep time.Duration
}

// NewChat initializes a chat handler and returns it
//
// d disables chat (bool)
// ls is the log symbol
// in is the channel for server -> discord
// out is the channel for discord -> server
func NewChat(d bool, ls string, cs storage.ChatsStore, in chan types.ChatMessage) func(w http.ResponseWriter, r *http.Request) {
	chat := Chat{d: d, ls: ls, cs: cs, in: in, sleep: 10 * time.Second}
	return chat.Handle
}

// Handle manages Rust <-> discord chat requests and logging
// Discord -> Rust is through the "out" chan and Rust -> Discord is
// through the "in" chan.
//
// HTTP POST requests are sent to the "in" chan
//
// HTTP GET requests wait for messages and disconnect with http.StatusNoContent
// after sleep seconds.
func (c *Chat) Handle(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	serverKey := context.Get(r, "serverKey").(string)
	account := context.Get(r, "account").(types.Account)

	if c.d {
		switch r.Method {
		case http.MethodPost:
			w.WriteHeader(http.StatusOK)
		case http.MethodGet:
			select {
			case <-time.After(c.sleep):
				break
			}
			w.WriteHeader(http.StatusNoContent)
		}
		return
	}

	switch r.Method {
	case http.MethodPost:
		decoder := json.NewDecoder(r.Body)
		var m types.ChatMessage
		err := decoder.Decode(&m)
		if err != nil {
			log.Println(c.ls + err.Error())
			return
		}

		m.Source = types.ChatSourceRust
		for _, s := range account.Servers {
			if s.Key == serverKey {
				m.ChannelID = s.ChatChanID
				c.in <- m
				return
			}
		}

	case http.MethodGet:
		var m types.ChatMessage
		err := c.cs.GetNext(serverKey, &m)
		if err != nil {
			if err.Error() != "not found" {
				log.Println(c.ls + err.Error())
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		b, err := json.Marshal(m)
		if err != nil {
			log.Println(c.ls + err.Error())
			return
		}
		// c.cs.Log(res)

		w.Write(b)
		// select {
		// case res := <-c.out:
		// 	b, err := json.Marshal(res)
		// 	if err != nil {
		// 		log.Println(c.ls + err.Error())
		// 		return
		// 	}
		// 	c.cs.Log(res)

		// 	w.Write(b)
		// case <-time.After(c.sleep):
		// 	w.WriteHeader(http.StatusNoContent)
		// }

	default:
		methodNotAllowed(w, c.ls)
	}
}
