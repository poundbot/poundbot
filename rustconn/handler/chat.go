package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"bitbucket.org/mrpoundsign/poundbot/db"

	"bitbucket.org/mrpoundsign/poundbot/types"
)

// A Chat is for handling discord <-> rust chat
type Chat struct {
	d     bool
	ls    string
	cs    db.ChatsStore
	in    chan string
	out   chan types.ChatMessage
	sleep time.Duration
}

// NewChat initializes a chat handler and returns it
//
// d disables chat (bool)
// ls is the log symbol
// in is the channel for server -> discord
// out is the channel for discord -> server
func NewChat(d bool, ls string, cs db.ChatsStore, in chan string, out chan types.ChatMessage) func(w http.ResponseWriter, r *http.Request) {
	chat := Chat{d: d, ls: ls, cs: cs, in: in, out: out, sleep: 10 * time.Second}
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
	// TODO: Make this less awful. Plugin must be updated to be smarter.
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
		var wg sync.WaitGroup
		decoder := json.NewDecoder(r.Body)
		var t types.ChatMessage
		err := decoder.Decode(&t)
		if err != nil {
			log.Println(c.ls + err.Error())
			return
		}

		t.Source = types.ChatSourceRust

		wg.Add(2)
		go func() {
			defer wg.Done()
			c.cs.Log(t)
		}()
		go func(t types.ChatMessage, c chan string) {
			defer wg.Done()
			var clan = ""
			if t.ClanTag != "" {
				clan = fmt.Sprintf("[%s] ", t.ClanTag)
			}
			c <- fmt.Sprintf("☢️ **%s%s**: %s", clan, t.DisplayName, t.Message)
		}(t, c.in)
		wg.Wait()

	case http.MethodGet:
		select {
		case res := <-c.out:
			b, err := json.Marshal(res)
			if err != nil {
				log.Println(c.ls + err.Error())
				return
			}
			c.cs.Log(res)

			w.Write(b)
		case <-time.After(c.sleep):
			w.WriteHeader(http.StatusNoContent)
		}

	default:
		methodNotAllowed(w, c.ls)
	}
}
