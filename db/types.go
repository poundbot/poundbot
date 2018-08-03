package db

import "mrpoundsign.com/poundbot/types"

// UsersAccessLayer is for accessing the user store.
//
// Get gets a user from store.
//
// BaseUpsert updates or creats a user in the store
//
// RemoveClan removes a clan tag from all users e.g. when a clan is removed.
//
// RemoveClansNotIn is used for removing all clan tags not in the provided
// list from all users in the data store.
//
// SetClanIn sets the clan tag on all users who have the provided steam IDs.
type UsersAccessLayer interface {
	Get(types.SteamInfo) (*types.User, error)
	BaseUpsert(types.BaseUser) error
	RemoveClan(tag string) error
	RemoveClansNotIn(tags []string) error
	SetClanIn(tag string, steamIds []uint64) error
}

// DiscordAuthsAccessLayer is for accessing the discord -> user authentications
// in the store.
//
// Upsert created or updates a discord auth
//
// Remove removes a discord auth
type DiscordAuthsAccessLayer interface {
	Upsert(types.DiscordAuth) error
	Remove(types.SteamInfo) error
}

// RaidAlertsAccessLayer is for accessing raid information. The raid information
// comes in as types.EntityDeath and comes out as types.RaidNotification
//
// GetReady gets raid alerts that are ready to alert
//
// AddInfo adds or updated raid information to a raid alert
//
// Remove deletes a raid alert
type RaidAlertsAccessLayer interface {
	GetReady(*[]types.RaidNotification) error
	AddInfo(types.EntityDeath) error
	Remove(types.RaidNotification) error
}

// ClansAccessLayer is for accessing clans data in the store
type ClansAccessLayer interface {
	Upsert(types.Clan) error
	Remove(tag string) error
	RemoveNotIn(tags []string) error
}

// ChatsAccessLayer is for logging chat
type ChatsAccessLayer interface {
	Log(types.ChatMessage) error
}

// DataAccessLayer is a complete implementation of the data store for users,
// clans, discord auth requests, raid alerts, and chats.
//
// Copy creates a new DB connection. Should always close the connection when
// you're done with it.
//
// Close closes the session
//
// CreateIndexes creates indexes, and should always be called when Poundbot
// first starts
type DataAccessLayer interface {
	Copy() DataAccessLayer
	Close()
	CreateIndexes()
	Users() UsersAccessLayer
	Chats() ChatsAccessLayer
	DiscordAuths() DiscordAuthsAccessLayer
	RaidAlerts() RaidAlertsAccessLayer
	Clans() ClansAccessLayer
}
