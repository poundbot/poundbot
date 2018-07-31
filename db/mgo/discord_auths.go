package mgo

import (
	"github.com/globalsign/mgo"
	"mrpoundsign.com/poundbot/db"
	"mrpoundsign.com/poundbot/types"
)

type DiscordAuths struct {
	collection *mgo.Collection
}

func (d DiscordAuths) Remove(si types.SteamInfo) error {
	return d.collection.Remove(si)
}

func (d DiscordAuths) Upsert(da types.DiscordAuth) error {
	_, err := d.collection.Upsert(da.SteamInfo, da)
	return err
}

var _ = db.DiscordAuthsAccessLayer(DiscordAuths{})
