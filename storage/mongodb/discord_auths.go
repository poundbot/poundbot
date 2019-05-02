package mongodb

import (
	"fmt"
	"context"

	"github.com/poundbot/poundbot/storage"
	"github.com/poundbot/poundbot/types"

	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

// A DiscordAuths implements db.DiscordAuthsStore
type DiscordAuths struct {
	collection mongo.Collection
}

func (d DiscordAuths) Get(discordName string) (types.DiscordAuth, error) {
	var da types.DiscordAuth
	result := d.collection.FindOne(context.Background(), bson.M{"discordname": discordName})
	err := result.Decode(&da)
	return da, err
}

func (d DiscordAuths) GetSnowflake(snowflake string) (types.DiscordAuth, error) {
	var da types.DiscordAuth
	result := d.collection.FindOne(context.Background(), bson.M{"snowflake": snowflake})
	err := result.Decode(&da)
	if err != nil {
		return types.DiscordAuth{}, fmt.Errorf("mongodb could not find snowflake %s (%s)", snowflake, err)
	}
	return da, nil
}

// Remove implements db.DiscordAuthsStore.Remove
func (d DiscordAuths) Remove(u storage.UserInfoGetter) error {
	_, err := d.collection.DeleteOne(context.Background(), bson.M{"playerid": u.GetPlayerID()})
	return err
}

// Upsert implements db.DiscordAuthsStore.Upsert
func (d DiscordAuths) Upsert(da types.DiscordAuth) error {
	upsert := true
	_, err := d.collection.UpdateOne(
		context.Background(),
		bson.M{"playerid": da.PlayerID},
		da,
		&options.UpdateOptions{Upsert: &upsert},
	)
	return err
}
