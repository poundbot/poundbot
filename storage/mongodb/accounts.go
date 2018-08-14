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

func (s Accounts) GetByDiscordGuild(key string, account *types.Account) error {
	return s.collection.Find(bson.M{accountsKeyField: key}).One(&account)
}

func (s Accounts) GetByServerKey(key string, account *types.Account) error {
	return s.collection.Find(bson.M{serverKeyField: key}).One(&account)
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

func (s Accounts) AddClan(key string, clan types.Clan) error {
	return s.collection.Update(
		bson.M{serverKeyField: key},
		bson.M{
			"$push": bson.M{"servers.$.clans": clan},
		},
	)
}

func (s Accounts) RemoveClan(key, clanTag string) error {
	return s.collection.Update(
		bson.M{serverKeyField: key, "servers.clans.tag": clanTag},
		bson.M{"$pull": bson.M{"servers.$.clans": bson.M{"tag": clanTag}}},
	)
}

func (s Accounts) SetClans(key string, clans []types.Clan) error {
	return s.collection.Update(
		bson.M{serverKeyField: key},
		bson.M{"$set": bson.M{"servers.$.clans": clans}},
	)
}
