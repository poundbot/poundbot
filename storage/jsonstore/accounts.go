package jsonstore

import (
	"bitbucket.org/mrpoundsign/poundbot/types"
)

type Accounts struct{}

func (s Accounts) All(*[]types.Account) error {
	return nil
}

func (s Accounts) GetByServerKey(serverKey string, serverAccount *types.Account) error {
	return nil
}

func (s Accounts) GetByDiscordGuild(snowflake string, serverAccount *types.Account) error {
	return nil
}

func (s Accounts) UpsertBase(types.BaseAccount) error {
	return nil
}

func (s Accounts) Remove(snowflake string) error {
	return nil
}

func (s Accounts) AddClan(serverKey string, clan types.Clan) error {
	return nil
}

func (s Accounts) RemoveClan(serverKey, clanTag string) error {
	return nil
}

func (s Accounts) SetClans(serverKey string, clans []types.Clan) error {
	return nil
}

func (s Accounts) AddServer(snowflake string, server types.Server) error {
	return nil
}

func (s Accounts) RemoveServer(snowflake, serverKey string) error {
	return nil
}

func (s Accounts) UpdateServer(snowflake string, server types.Server) error {
	return nil
}

func (s Accounts) RemoveNotInDiscordGuildList(guildIDs []string) error {
	return nil
}
