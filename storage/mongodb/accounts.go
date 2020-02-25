package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/poundbot/poundbot/pbclock"
	"github.com/poundbot/poundbot/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const accountsKeyField = "guildsnowflake"
const serverKeyField = "servers.key"

var iclock = pbclock.Clock

type Accounts struct {
	collection *mongo.Collection
}

func (s Accounts) All(accounts *[]types.Account) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	curs, err := s.collection.Find(ctx, bson.M{})
	if err != nil {
		return err
	}

	return curs.All(ctx, accounts)
}

func (s Accounts) GetByDiscordGuild(key string) (types.Account, error) {
	var account types.Account
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := s.collection.FindOne(ctx, bson.M{accountsKeyField: key}).Decode(&account)
	return account, err
}

func (s Accounts) GetByServerKey(key string) (types.Account, error) {
	var account types.Account
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := s.collection.FindOne(ctx, bson.M{serverKeyField: key}).Decode(&account)
	return account, err
}

func (s Accounts) UpsertBase(account types.BaseAccount) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.collection.UpdateOne(
		ctx,
		bson.M{accountsKeyField: account.GuildSnowflake},
		bson.M{
			"$setOnInsert": types.NewTimestamp(),
			"$set":         account,
		},
		options.Update().SetUpsert(true),
	)
	return err
}

func (s Accounts) Remove(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	u, err := s.collection.UpdateOne(
		ctx,
		bson.M{accountsKeyField: key},
		bson.M{"$set": bson.M{"disabled": true}},
	)
	if u.MatchedCount != 1 {
		return errors.New("could not find account to remove")
	}
	return err
}

func (s Accounts) AddClan(serverKey string, clan types.Clan) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.collection.UpdateOne(
		ctx,
		bson.M{serverKeyField: serverKey},
		bson.M{
			"$push": bson.M{"servers.$.clans": clan},
		},
	)
	return err
}

func (s Accounts) RemoveClan(serverKey, clanTag string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.collection.UpdateOne(
		ctx,
		bson.M{serverKeyField: serverKey, "servers.clans.tag": clanTag},
		bson.M{"$pull": bson.M{"servers.$.clans": bson.M{"tag": clanTag}}},
	)
	return err
}

func (s Accounts) SetClans(serverKey string, clans []types.Clan) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	u, err := s.collection.UpdateOne(
		ctx,
		bson.M{serverKeyField: serverKey},
		bson.M{"$set": bson.M{"servers.$.clans": clans}},
	)
	if u.MatchedCount != 1 {
		return errors.New("nothing matched")
	}
	return err
}

func (s Accounts) AddServer(snowflake string, server types.AccountServer) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.CreatedAt = iclock().Now().UTC()
	_, err := s.collection.UpdateOne(
		ctx,
		bson.M{accountsKeyField: snowflake},
		bson.M{"$push": bson.M{"servers": server}},
	)
	return err
}

func (s Accounts) RemoveServer(snowflake, serverKey string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.collection.UpdateOne(
		ctx,
		bson.M{serverKeyField: serverKey},
		bson.M{"$pull": bson.M{"servers": bson.M{"key": serverKey}}},
	)
	return err
}

func (s Accounts) UpdateServer(snowflake, oldKey string, server types.AccountServer) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.collection.UpdateOne(
		ctx,
		bson.M{
			accountsKeyField: snowflake,
			serverKeyField:   oldKey,
		},
		bson.M{"$set": bson.M{"servers.$": server}},
	)
	return err
}

func (s Accounts) RemoveNotInDiscordGuildList(guilds []types.BaseAccount) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	insertTS := types.NewTimestamp()
	insertTS.CreatedAt = iclock().Now().UTC()
	guildIDs := make([]string, len(guilds))

	type doc struct {
		types.BaseAccount `bson:",inline"`
		Disabled          bool
		UpdatedAt         time.Time
	}

	for i, guild := range guilds {
		// Collect the IDs for disabling later
		guildIDs[i] = guild.GuildSnowflake

		_, err := s.collection.UpdateOne(
			ctx,
			bson.M{
				accountsKeyField: guild.GuildSnowflake,
			},
			bson.M{
				"$setOnInsert": bson.M{"createdat": insertTS.CreatedAt},
				"$set":         doc{BaseAccount: guild, Disabled: false, UpdatedAt: insertTS.UpdatedAt},
			},
			options.Update().SetUpsert(true),
		)
		if err != nil {
			return fmt.Errorf("error updating guild: %s", err)
		}
	}

	// Now disable all the guilds not in the list
	_, err := s.collection.UpdateMany(
		ctx,
		bson.M{
			accountsKeyField: bson.M{"$nin": guildIDs},
		},
		bson.M{"$set": bson.M{"disabled": true}},
	)

	return err
}

func (s Accounts) SetRegisteredPlayerIDs(accoutID string, playerIDs []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.collection.UpdateOne(
		ctx,
		bson.M{accountsKeyField: accoutID},
		bson.M{
			"$set": bson.M{
				"registeredplayerids": playerIDs,
			},
		},
	)
	return err
}

func (s Accounts) AddRegisteredPlayerIDs(accoutID string, playerIDs []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.collection.UpdateOne(
		ctx,
		bson.M{accountsKeyField: accoutID},
		bson.M{
			"$addToSet": bson.M{
				"registeredplayerids": bson.M{
					"$each": playerIDs,
				},
			},
		},
	)
	return err
}

func (s Accounts) RemoveRegisteredPlayerIDs(accoutID string, playerIDs []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.collection.UpdateOne(
		ctx,
		bson.M{accountsKeyField: accoutID},
		bson.M{"$pullAll": bson.M{"registeredplayerids": playerIDs}},
	)
	return err
}

func (s Accounts) Touch(serverKey string) error {
	now := iclock().Now().UTC()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.collection.UpdateOne(
		ctx,
		bson.M{serverKeyField: serverKey},
		bson.M{
			"$set": bson.M{
				"updatedat":           now,
				"servers.$.updatedat": now,
			},
		},
	)
	return err
}
