package jsonstore

import (
	"fmt"
	"time"

	"bitbucket.org/mrpoundsign/poundbot/types"
)

// A Users implements db.UsersStore
type Users struct{}

// Get implements db.UsersStore.Get
func (u Users) Get(steamID uint64, user *types.User) error {
	err := users.s.driver.Read(
		users.s.collection,
		fmt.Sprintf("%d", steamID),
		&user,
	)
	return err
}

// UpsertBase implements db.UsersStore.UpsertBase
func (u Users) UpsertBase(base types.BaseUser) error {
	var user types.User
	// TODO: Preserve things
	err := u.Get(base.SteamID, &user)
	if err != nil {
		user.CreatedAt = time.Now().UTC()
	}
	user.BaseUser = base

	if err := users.s.driver.Write(
		users.s.collection,
		fmt.Sprintf("%d", user.SteamID),
		&user,
	); err != nil {
		return err
	}

	return nil
}

// RemoveClan implements db.UsersStore.RemoveClan
func (u Users) RemoveClan(serverKey, tag string) error {
	return nil
}

// RemoveClansNotIn implements db.UsersStore.RemoveClansNotIn
func (u Users) RemoveClansNotIn(serverKey string, tags []string) error {
	return nil
}

// SetClanIn implements db.UsersStore.SetClanIn
func (u Users) SetClanIn(serverKey, tag string, steamIds []uint64) error {
	return nil
}
