package mongodb

import (
	"fmt"

	"github.com/poundbot/poundbot/types"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// A DiscordAuths implements db.DiscordAuthsStore
type DiscordAuths struct {
	collection *mgo.Collection
}

func (d DiscordAuths) Get(discordName string) (types.DiscordAuth, error) {
	var da types.DiscordAuth
	err := d.collection.Find(bson.M{"discordname": discordName}).One(&da)
	return da, err
}

func (d DiscordAuths) GetSnowflake(snowflake string) (types.DiscordAuth, error) {
	var da types.DiscordAuth
	err := d.collection.Find(bson.M{"snowflake": snowflake}).One(&da)
	if err != nil {
		return types.DiscordAuth{}, fmt.Errorf("mongodb could not find snowflake %s (%s)", snowflake, err)
	}
	return da, nil
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
