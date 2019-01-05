package mongodb

import (
	"bitbucket.org/mrpoundsign/poundbot/types"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// A DiscordAuths implements db.DiscordAuthsStore
type DiscordAuths struct {
	collection *mgo.Collection
}

func (d DiscordAuths) Get(discordName string, da *types.DiscordAuth) error {
	return d.collection.Find(bson.M{"discordname": discordName}).One(&da)
}

func (d DiscordAuths) GetSnowflake(snowflake string, da *types.DiscordAuth) error {
	return d.collection.Find(bson.M{"snowflake": snowflake}).One(&da)
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
