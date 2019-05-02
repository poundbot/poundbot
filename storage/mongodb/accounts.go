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

const (
	accountKeyField      = "guildsnowflake"
	accountDisabledField = "disabled"
	accountServersField  = "servers"

	serverKeyField = accountServersField + ".key"
	serverSelector = accountServersField + ".$"

	clanSelector = serverSelector + ".clans"
	clanTagField = accountServersField + ".clans.tag"
)

var iclock = pbclock.Clock

type Accounts struct {
	collection mongo.Collection
}

func (s Accounts) All(accounts *[]types.Account) error {
	cur, err := s.collection.Find(context.Background(), bson.M{})
	if err != nil {
		return err
	}
	defer cur.Close(context.Background())

	for cur.Next(context.Background()) {
		var doc types.Account
		err := cur.Decode(&doc)
		if err != nil {
			return err
		}
		*accounts = append(*accounts, doc)
	}

	return nil
	// return .All(accounts)
}

func (s Accounts) GetByDiscordGuild(key string) (types.Account, error) {
	var account types.Account
	result := s.collection.FindOne(context.Background(), bson.M{accountKeyField: key})
	err := result.Decode(&account)
	return account, err
}

func (s Accounts) GetByServerKey(key string) (types.Account, error) {
	var account types.Account
	result := s.collection.FindOne(context.Background(), bson.M{serverKeyField: key})
	err := result.Decode(&account)
	return account, err
}

func (s Accounts) UpsertBase(account types.BaseAccount) error {
	upsert := true
	_, err := s.collection.UpdateOne(
		context.Background(),
		bson.M{accountKeyField: account.GuildSnowflake},
		bson.M{
			"$setOnInsert": types.NewTimestamp(),
			"$set":         account,
		},
		&options.UpdateOptions{Upsert: &upsert},
	)
	return err
}

func (s Accounts) Remove(key string) error {
	dr, err := s.collection.DeleteOne(context.Background(), bson.M{accountKeyField: key})
	if dr.DeletedCount != 1 {
		return errors.New("not found")
	}
	return err
}

func (s Accounts) AddClan(serverKey string, clan types.Clan) error {
	upsert := true
	_, err := s.collection.UpdateOne(
		context.Background(),
		bson.M{serverKeyField: serverKey},
		bson.M{
			"$push": bson.M{clanSelector: clan},
		},
		&options.UpdateOptions{Upsert: &upsert},
	)
	return err
}

func (s Accounts) RemoveClan(serverKey, clanTag string) error {
	ur, err := s.collection.UpdateOne(
		context.Background(),
		bson.M{serverKeyField: serverKey, clanTagField: clanTag},
		bson.M{"$pull": bson.M{clanSelector: bson.M{"tag": clanTag}}},
	)
	if ur.ModifiedCount != 1 {
		return errors.New("not found")
	}
	return err
}

func (s Accounts) SetClans(serverKey string, clans []types.Clan) error {
	ur, err := s.collection.UpdateOne(
		context.Background(),
		bson.M{serverKeyField: serverKey},
		bson.M{"$set": bson.M{clanSelector: clans}},
	)
	if ur.ModifiedCount != 1 {
		return errors.New("not found")
	}
	return err
}

func (s Accounts) AddServer(snowflake string, server types.Server) error {
	server.CreatedAt = iclock().Now().UTC()
	_, err := s.collection.UpdateOne(
		context.Background(),
		bson.M{accountKeyField: snowflake},
		bson.M{"$push": bson.M{accountServersField: server}},
	)
	return err
}

func (s Accounts) RemoveServer(snowflake, serverKey string) error {
	ur, err := s.collection.UpdateOne(
		context.Background(),
		bson.M{serverKeyField: serverKey},
		bson.M{"$pull": bson.M{accountServersField: bson.M{"key": serverKey}}},
	)
	if ur.ModifiedCount != 1 {

	}
	return err
}

func (s Accounts) UpdateServer(snowflake, oldKey string, server types.Server) error {
	_, err := s.collection.UpdateOne(
		context.Background(),
		bson.M{
			accountKeyField: snowflake,
			serverKeyField:  oldKey,
		},
		bson.M{"$set": bson.M{serverSelector: server}},
	)
	return err
}

func (s Accounts) RemoveNotInDiscordGuildList(guilds []types.BaseAccount) error {
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
			context.Background(),
			bson.M{
				accountKeyField: guild.GuildSnowflake,
			},
			bson.M{
				"$setOnInsert": bson.M{"createdat": insertTS.CreatedAt},
				"$set":         doc{BaseAccount: guild, Disabled: false, UpdatedAt: insertTS.UpdatedAt},
			},
			upsertOptions(),
		)
		if err != nil {
			return fmt.Errorf("error updating guild: %s", err)
		}
	}

	// Now disable all the guilds not in the list
	_, err := s.collection.UpdateMany(
		context.Background(),
		bson.M{
			accountKeyField: bson.M{"$nin": guildIDs},
		},
		bson.M{"$set": bson.M{accountDisabledField: true}},
	)

	return err
}

func (s Accounts) Touch(serverKey string) error {
	now := iclock().Now().UTC()
	_, err := s.collection.UpdateOne(
		context.Background(),
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
