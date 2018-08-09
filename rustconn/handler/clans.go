package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"bitbucket.org/mrpoundsign/poundbot/db"
	"bitbucket.org/mrpoundsign/poundbot/types"
)

type Clans struct {
	ls string
	cs db.ClansStore
	us db.UsersStore
}

func NewClans(ls string, cs db.ClansStore, us db.UsersStore) func(w http.ResponseWriter, r *http.Request) {
	clans := Clans{ls, cs, us}
	return clans.Handle
}

// Handle manages clans sync HTTP requests from the Rust server
// These requests are a complete refresh of all clans
func (c *Clans) Handle(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var t []types.ServerClan
	err := decoder.Decode(&t)
	if err != nil {
		log.Println(c.ls + err.Error())
		return
	}

	clanCount := len(t)
	clans := make([]types.Clan, clanCount)
	tags := make([]string, clanCount)
	for i, sc := range t {
		cl, err := types.ClanFromServerClan(sc)
		if err != nil {
			log.Printf(c.ls+"clansHandler Error: %v\n", err)
			handleError(w, c.ls, types.RESTError{
				StatusCode: http.StatusBadRequest,
				Error:      "Error processing clan data",
			})
			return
		}
		tags[i] = cl.Tag
		clans[i] = *cl
	}

	for _, clan := range clans {
		c.cs.Upsert(clan)
	}

	c.cs.RemoveNotIn(tags)
	c.us.RemoveClansNotIn(tags)
}
