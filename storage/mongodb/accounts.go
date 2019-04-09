package mongodb

import (
	"time"

	"bitbucket.org/mrpoundsign/poundbot/types"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

const accountsKeyField = "guildsnowflake"
const serverKeyField = "servers.key"

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
			"$setOnInsert": bson.M{"createdat": time.Now().UTC()},
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

func (s Accounts) SetClans(key string, clans []types.Clan) error {
	return s.collection.Update(
		bson.M{serverKeyField: key},
		bson.M{"$set": bson.M{"servers.$.clans": clans}},
	)
}

func (s Accounts) AddServer(snowflake string, server types.Server) error {
	return s.collection.Update(
		bson.M{accountsKeyField: snowflake},
		bson.M{
			"$push": bson.M{"servers": server},
		},
	)
}

func (s Accounts) RemoveServer(snowflake, serverKey string) error {
	return s.collection.Update(
		bson.M{serverKeyField: serverKey},
		bson.M{"$pull": bson.M{"servers": bson.M{"key": serverKey}}},
	)
}

func (s Accounts) UpdateServer(snowflake, oldKey string, server types.Server) error {
	return s.collection.Update(
		bson.M{
			accountsKeyField: snowflake,
			serverKeyField:   oldKey,
		},
		bson.M{"$set": bson.M{"servers.$": server}},
	)
}

func (s Accounts) RemoveNotInDiscordGuildList(guildIDs []string) error {
	_, err := s.collection.UpdateAll(
		bson.M{
			accountsKeyField: bson.M{
				"$nin": guildIDs,
			},
		},
		bson.M{"$set": bson.M{"disabled": true}},
	)

	if err != nil {
		return err
	}

	_, err = s.collection.UpdateAll(
		bson.M{
			accountsKeyField: bson.M{
				"$in": guildIDs,
			},
		},
		bson.M{"$set": bson.M{"disabled": false}},
	)
	return err
}
