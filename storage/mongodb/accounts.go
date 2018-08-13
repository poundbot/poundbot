package mongodb

import (
	"bitbucket.org/mrpoundsign/poundbot/types"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type Accounts struct {
	collection *mgo.Collection
}

func (s Accounts) All(*[]types.Account) error {
	return nil
}

func (s Accounts) GetByDiscordGuild(key string, account *types.Account) error {
	return s.collection.Find(bson.M{"accountdiscord.guildsnowflake": key}).One(&account)
}

func (s Accounts) GetByServerKey(key string, account *types.Account) error {
	return s.collection.Find(bson.M{"servers.key": key}).One(&account)
}

func (s Accounts) UpsertBase(types.BaseAccount) error {
	return nil
}

func (s Accounts) Remove(key string) error {
	return nil
}

func (s Accounts) AddClan(key string, clan types.Clan) error {
	return nil
}

func (s Accounts) RemoveClan(key, clanTag string) error {
	return nil
}

func (s Accounts) SetClans(key string, clans []types.Clan) error {
	return s.collection.Update(
		bson.M{"servers.key": key},
		bson.M{"$set": bson.M{"servers.$.clans": clans}},
	)
}
