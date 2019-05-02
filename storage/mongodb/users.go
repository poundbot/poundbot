package mongodb

import (
	"time"
	"context"

	"github.com/poundbot/poundbot/storage"
	"github.com/poundbot/poundbot/types"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// A Users implements db.UsersStore
type Users struct {
	collection mongo.Collection
}

// Get implements db.UsersStore.Get
func (u Users) Get(gameUserID string) (types.User, error) {
	var user types.User
	result := u.collection.FindOne(context.Background(), bson.M{"playerids": gameUserID})
	err := result.Decode(&user)
	return user, err
}

func (u Users) GetSnowflake(snowflake string) (types.User, error) {
	var user types.User
	result := u.collection.FindOne(context.Background(), bson.M{"snowflake": snowflake})
	err := result.Decode(&user)
	return user, err
}

func (u Users) UpsertPlayer(info storage.UserInfoGetter) error {
	upsert := true
	_, err := u.collection.UpdateOne(
		context.Background(),
		bson.M{"snowflake": info.GetDiscordID()},
		bson.M{
			"$setOnInsert": bson.M{
				"snowflake": info.GetDiscordID(),
				"createdat": time.Now().UTC(),
			},
			"$addToSet": bson.M{"playerids": info.GetPlayerID()},
		},
		&options.UpdateOptions{Upsert: &upsert},
	)

	return err
}
