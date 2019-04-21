package storage

import (
	"time"

	"github.com/poundbot/poundbot/types"
)

type UserInfoGetter interface {
	GetPlayerID() string
	GetDiscordID() string
}

// UsersStore is for accessing the user store.
//
// Get gets a user from store.
//
// UpsertBase updates or creates a user in the store
//
// RemoveClan removes a clan tag from all users e.g. when a clan is removed.
//
// RemoveClansNotIn is used for removing all clan tags not in the provided
// list from all users in the data store.
//
// SetClanIn sets the clan tag on all users who have the provided steam IDs.
type UsersStore interface {
	Get(PlayerID string) (types.User, error)
	GetSnowflake(snowflake string) (types.User, error)
	UpsertPlayer(info UserInfoGetter) error
}

// DiscordAuthsStore is for accessing the discord -> user authentications
// in the store.
//
// Upsert created or updates a discord auth
//
// Remove removes a discord auth
type DiscordAuthsStore interface {
	Get(discordName string) (types.DiscordAuth, error)
	GetSnowflake(snowflake string) (types.DiscordAuth, error)
	Upsert(types.DiscordAuth) error
	Remove(UserInfoGetter) error
}

// RaidAlertsStore is for accessing raid information. The raid information
// comes in as types.EntityDeath and comes out as types.RaidAlert
//
// GetReady gets raid alerts that are ready to alert
//
// AddInfo adds or updated raid information to a raid alert
//
// Remove deletes a raid alert
type RaidAlertsStore interface {
	GetReady() ([]types.RaidAlert, error)
	AddInfo(alertIn time.Duration, ed types.EntityDeath) error
	Remove(types.RaidAlert) error
}

// AccountsStore is for accounts storage
type AccountsStore interface {
	All(*[]types.Account) error
	GetByDiscordGuild(snowflake string) (types.Account, error)
	GetByServerKey(serverKey string) (types.Account, error)
	UpsertBase(types.BaseAccount) error
	Remove(snowflake string) error

	AddServer(snowflake string, server types.Server) error
	UpdateServer(snowflake, oldKey string, server types.Server) error
	RemoveServer(snowflake, serverKey string) error

	AddClan(serverKey string, clan types.Clan) error
	RemoveClan(serverKey, clanTag string) error
	SetClans(serverKey string, clans []types.Clan) error

	RemoveNotInDiscordGuildList(guildIDs []types.BaseAccount) error
	Touch(serverKey string) error
}

// Storage is a complete implementation of the data store for users,
// clans, discord auth requests, raid alerts, and chats.
//
// Copy creates a new DB connection. Should always close the connection when
// you're done with it.
//
// Close closes the session
//
// Init creates indexes, and should always be called when Poundbot
// first starts
type Storage interface {
	Copy() Storage
	Close()
	Init()
	Accounts() AccountsStore
	Users() UsersStore
	DiscordAuths() DiscordAuthsStore
	RaidAlerts() RaidAlertsStore
}
