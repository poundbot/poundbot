package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/poundbot/poundbot/storage"
	"github.com/poundbot/poundbot/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// A DiscordAuths implements db.DiscordAuthsStore
type DiscordAuths struct {
	collection *mongo.Collection
}

func (d DiscordAuths) GetByDiscordName(discordName string) (types.DiscordAuth, error) {
	var da types.DiscordAuth
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := d.collection.FindOne(ctx, bson.M{"discordname": discordName}).Decode(&da)
	return da, err
}

func (d DiscordAuths) GetByDiscordID(snowflake string) (types.DiscordAuth, error) {
	var da types.DiscordAuth
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := d.collection.FindOne(ctx, bson.M{"snowflake": snowflake}).Decode(&da)
	if err != nil {
		return types.DiscordAuth{}, fmt.Errorf("mongodb could not find snowflake %s (%s)", snowflake, err)
	}
	return da, nil
}

// Remove implements db.DiscordAuthsStore.Remove
func (d DiscordAuths) Remove(u storage.UserInfoGetter) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res, err := d.collection.DeleteMany(ctx, bson.M{"playerid": u.GetPlayerID()})
	if res.DeletedCount == 0 {
		log.WithField("sys", "DiscordAuths").Tracef("No auths removed for %s", u.GetPlayerID())
	}
	return err
}

// Upsert implements db.DiscordAuthsStore.Upsert
func (d DiscordAuths) Upsert(da types.DiscordAuth) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := d.collection.UpdateOne(
		ctx,
		bson.M{"playerid": da.PlayerID},
		da,
		options.Update().SetUpsert(true),
	)
	return err
}
