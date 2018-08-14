package storage

import (
	"time"

	"bitbucket.org/mrpoundsign/poundbot/types"
)

// UsersStore is for accessing the user store.
//
// Get gets a user from store.
//
// UpsertBase updates or creats a user in the store
//
// RemoveClan removes a clan tag from all users e.g. when a clan is removed.
//
// RemoveClansNotIn is used for removing all clan tags not in the provided
// list from all users in the data store.
//
// SetClanIn sets the clan tag on all users who have the provided steam IDs.
type UsersStore interface {
	Get(steamID uint64, u *types.User) error
	UpsertBase(baseUser types.BaseUser) error
}

// DiscordAuthsStore is for accessing the discord -> user authentications
// in the store.
//
// Upsert created or updates a discord auth
//
// Remove removes a discord auth
type DiscordAuthsStore interface {
	Get(discordName string, da *types.DiscordAuth) error
	GetSnowflake(snowflake string, da *types.DiscordAuth) error
	Upsert(types.DiscordAuth) error
	Remove(types.SteamInfo) error
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
	GetReady(*[]types.RaidAlert) error
	AddInfo(alertIn time.Duration, ed types.EntityDeath) error
	Remove(types.RaidAlert) error
}

// ChatsStore is for logging chat
type ChatsStore interface {
	Log(types.ChatMessage) error
	GetNext(serverKey string, ChatMessage *types.ChatMessage) error
}

// AccountsStore is for accounts storage
type AccountsStore interface {
	All(*[]types.Account) error
	GetByDiscordGuild(key string, account *types.Account) error
	GetByServerKey(key string, account *types.Account) error
	UpsertBase(types.BaseAccount) error
	Remove(key string) error
	AddClan(key string, clan types.Clan) error
	RemoveClan(key, clanTag string) error
	SetClans(key string, clans []types.Clan) error
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
	Chats() ChatsStore
	DiscordAuths() DiscordAuthsStore
	RaidAlerts() RaidAlertsStore
}
