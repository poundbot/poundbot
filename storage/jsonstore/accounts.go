package jsonstore

import (
	"bitbucket.org/mrpoundsign/poundbot/types"
)

type Accounts struct{}

func (s Accounts) All(*[]types.Account) error {
	return nil
}

func (s Accounts) GetByServerKey(key string, serverAccount *types.Account) error {
	return nil
}

func (s Accounts) GetByDiscordGuild(key string, serverAccount *types.Account) error {
	return nil
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

func (s Accounts) RemoveClan(clanTag, key string) error {
	return nil
}

func (s Accounts) SetClans(key string, clans []types.Clan) error {
	return nil
}
