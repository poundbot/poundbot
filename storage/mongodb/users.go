package mongodb

import (
	"context"
	"time"

	"github.com/poundbot/poundbot/storage"
	"github.com/poundbot/poundbot/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const userPlayerIDsField = "playerids"
const userSnowflakeField = "snowflake"

// A Users implements db.UsersStore
type Users struct {
	collection *mongo.Collection
}

// Get implements db.UsersStore.Get
func (u Users) GetByPlayerID(gameUserID string) (types.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var user types.User
	err := u.collection.FindOne(ctx, bson.M{userPlayerIDsField: gameUserID}).Decode(&user)
	return user, err
}

func (u Users) GetByDiscordID(snowflake string) (types.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var user types.User
	err := u.collection.FindOne(ctx, bson.M{userSnowflakeField: snowflake}).Decode(&user)
	return user, err
}

func (u Users) GetPlayerIDsByDiscordIDs(snowflakes []string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ids, err := u.collection.Distinct(ctx, userPlayerIDsField, bson.M{userSnowflakeField: bson.M{"$in": snowflakes}})
	playerIDs := make([]string, len(ids))
	for i, id := range ids {
		playerIDs[i] = id.(string)
	}
	return playerIDs, err
}

func (u Users) UpsertPlayer(info storage.UserInfoGetter) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := u.collection.UpdateOne(
		ctx,
		bson.M{userSnowflakeField: info.GetDiscordID()},
		bson.M{
			"$setOnInsert": bson.M{
				userSnowflakeField: info.GetDiscordID(),
				"createdat":        time.Now().UTC(),
			},
			"$set":      bson.M{"updatedat": time.Now().UTC()},
			"$addToSet": bson.M{userPlayerIDsField: info.GetPlayerID()},
		},
		options.Update().SetUpsert(true),
	)

	return err
}

func (u Users) RemovePlayerID(snowflake, playerID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if playerID == "all" {
		_, err := u.collection.DeleteOne(
			ctx,
			bson.M{userSnowflakeField: snowflake},
		)
		return err
	}

	_, err := u.collection.UpdateOne(
		ctx,
		bson.M{userSnowflakeField: snowflake},
		bson.M{
			"$set": bson.M{
				"updatedat": time.Now().UTC(),
			},
			"$pull": bson.M{userPlayerIDsField: playerID},
		},
	)
	return err
}
