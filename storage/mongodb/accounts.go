package mongodb

import (
	"fmt"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/poundbot/poundbot/pbclock"
	"github.com/poundbot/poundbot/types"
)

const accountsKeyField = "guildsnowflake"
const serverKeyField = "servers.key"

var iclock = pbclock.Clock

type Accounts struct {
	collection *mgo.Collection
}

func (s Accounts) All(accounts *[]types.Account) error {
	return s.collection.Find(bson.M{}).All(accounts)
}

func (s Accounts) GetByDiscordGuild(key string) (types.Account, error) {
	var account types.Account
	err := s.collection.Find(bson.M{accountsKeyField: key}).One(&account)
	return account, err
}

func (s Accounts) GetByServerKey(key string) (types.Account, error) {
	var account types.Account
	err := s.collection.Find(bson.M{serverKeyField: key}).One(&account)
	return account, err
}

func (s Accounts) UpsertBase(account types.BaseAccount) error {
	_, err := s.collection.Upsert(
		bson.M{accountsKeyField: account.GuildSnowflake},
		bson.M{
			"$setOnInsert": types.NewTimestamp(),
			"$set":         account,
		},
	)
	return err
}

func (s Accounts) Remove(key string) error {
	return s.collection.Remove(bson.M{accountsKeyField: key})
}

func (s Accounts) AddClan(serverKey string, clan types.Clan) error {
	return s.collection.Update(
		bson.M{serverKeyField: serverKey},
		bson.M{
			"$push": bson.M{"servers.$.clans": clan},
		},
	)
}

func (s Accounts) RemoveClan(serverKey, clanTag string) error {
	return s.collection.Update(
		bson.M{serverKeyField: serverKey, "servers.clans.tag": clanTag},
		bson.M{"$pull": bson.M{"servers.$.clans": bson.M{"tag": clanTag}}},
	)
}

func (s Accounts) SetClans(serverKey string, clans []types.Clan) error {
	return s.collection.Update(
		bson.M{serverKeyField: serverKey},
		bson.M{"$set": bson.M{"servers.$.clans": clans}},
	)
}

func (s Accounts) AddServer(snowflake string, server types.AccountServer) error {
	server.CreatedAt = iclock().Now().UTC()
	return s.collection.Update(
		bson.M{accountsKeyField: snowflake},
		bson.M{"$push": bson.M{"servers": server}},
	)
}

func (s Accounts) RemoveServer(snowflake, serverKey string) error {
	return s.collection.Update(
		bson.M{serverKeyField: serverKey},
		bson.M{"$pull": bson.M{"servers": bson.M{"key": serverKey}}},
	)
}

func (s Accounts) UpdateServer(snowflake, oldKey string, server types.AccountServer) error {
	return s.collection.Update(
		bson.M{
			accountsKeyField: snowflake,
			serverKeyField:   oldKey,
		},
		bson.M{"$set": bson.M{"servers.$": server}},
	)
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

		_, err := s.collection.Upsert(
			bson.M{
				accountsKeyField: guild.GuildSnowflake,
			},
			bson.M{
				"$setOnInsert": bson.M{"createdat": insertTS.CreatedAt},
				"$set":         doc{BaseAccount: guild, Disabled: false, UpdatedAt: insertTS.UpdatedAt},
			},
		)
		if err != nil {
			return fmt.Errorf("error updating guild: %s", err)
		}
	}

	// Now disable all the guilds not in the list
	_, err := s.collection.UpdateAll(
		bson.M{
			accountsKeyField: bson.M{"$nin": guildIDs},
		},
		bson.M{"$set": bson.M{"disabled": true}},
	)

	return err
}

func (s Accounts) SetRegisteredPlayerIDs(accoutID string, playerIDs []string) error {
	return s.collection.Update(
		bson.M{accountsKeyField: accoutID},
		bson.M{
			"$set": bson.M{
				"registeredplayerids": playerIDs,
			},
		},
	)
}

func (s Accounts) AddRegisteredPlayerIDs(accoutID string, playerIDs []string) error {
	return s.collection.Update(
		bson.M{accountsKeyField: accoutID},
		bson.M{
			"$addToSet": bson.M{
				"registeredplayerids": bson.M{
					"$each": playerIDs,
				},
			},
		},
	)
}

func (s Accounts) RemoveRegisteredPlayerIDs(accoutID string, playerIDs []string) error {
	return s.collection.Update(
		bson.M{accountsKeyField: accoutID},
		bson.M{"$pullAll": bson.M{"registeredplayerids": playerIDs}},
	)
}

func (s Accounts) Touch(serverKey string) error {
	now := iclock().Now().UTC()
	return s.collection.Update(
		bson.M{serverKeyField: serverKey},
		bson.M{
			"$set": bson.M{
				"updatedat":           now,
				"servers.$.updatedat": now,
			},
		},
	)
}
