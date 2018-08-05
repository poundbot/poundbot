package mongodb

import (
	"bitbucket.org/mrpoundsign/poundbot/types"
	"github.com/globalsign/mgo"
)

// A DiscordAuths implements db.DiscordAuthsStore
type DiscordAuths struct {
	collection *mgo.Collection
}

// Remove implements db.DiscordAuthsStore.Remove
func (d DiscordAuths) Remove(si types.SteamInfo) error {
	return d.collection.Remove(si)
}

// Upsert implements db.DiscordAuthsStore.Upsert
func (d DiscordAuths) Upsert(da types.DiscordAuth) error {
	_, err := d.collection.Upsert(da.SteamInfo, da)
	return err
}
