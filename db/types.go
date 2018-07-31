package db

import "mrpoundsign.com/poundbot/types"

type UsersAccessLayer interface {
	Get(types.SteamInfo) (*types.User, error)
	BaseUpsert(types.BaseUser) error
	RemoveClan(tag string) error
	RemoveClansNotIn(tags []string) error
	SetClanIn(tag string, steam_ids []uint64) error
}

type DiscordAuthsAccessLayer interface {
	Upsert(types.DiscordAuth) error
	Remove(types.SteamInfo) error
}

type RaidAlertsAccessLayer interface {
	GetReady(*[]types.RaidNotification) error
	AddInfo(types.EntityDeath) error
	Remove(types.RaidNotification) error
}

type ClansAccessLayer interface {
	Upsert(types.Clan) error
	Remove(tag string) error
	RemoveNotIn(tags []string) error
}

type ChatsAccessLayer interface {
	Log(types.ChatMessage) error
}

type DataAccessLayer interface {
	// Copy creates a new DB connection
	Copy() DataAccessLayer
	// Close closes the session
	Close()
	CreateIndexes()
	Users() UsersAccessLayer
	Chats() ChatsAccessLayer
	DiscordAuths() DiscordAuthsAccessLayer
	RaidAlerts() RaidAlertsAccessLayer
	Clans() ClansAccessLayer
}
