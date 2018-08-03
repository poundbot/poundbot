package mgo

import (
	"github.com/globalsign/mgo"
	"mrpoundsign.com/poundbot/types"
)

// A DiscordAuths implements db.DiscordAuthsAccessLayer
type DiscordAuths struct {
	collection *mgo.Collection
}

// Remove implements db.DiscordAuthsAccessLayer.Remove
func (d DiscordAuths) Remove(si types.SteamInfo) error {
	return d.collection.Remove(si)
}

// Upsert implements db.DiscordAuthsAccessLayer.Upsert
func (d DiscordAuths) Upsert(da types.DiscordAuth) error {
	_, err := d.collection.Upsert(da.SteamInfo, da)
	return err
}
