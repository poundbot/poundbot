package jsonstore

import (
	"fmt"
	"time"

	"bitbucket.org/mrpoundsign/poundbot/types"
)

// A Users implements db.UsersStore
type Users struct{}

// Get implements db.UsersStore.Get
func (u Users) Get(si types.SteamInfo) (*types.User, error) {
	var user types.User
	err := users.s.driver.Read(
		users.s.collection,
		fmt.Sprintf("%d", si.SteamID),
		&user,
	)
	return &user, err
}

// UpsertBase implements db.UsersStore.UpsertBase
func (u Users) UpsertBase(base types.BaseUser) error {
	// TODO: Preserve things
	user, err := u.Get(base.SteamInfo)
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
func (u Users) RemoveClan(tag string) error {
	return nil
}

// RemoveClansNotIn implements db.UsersStore.RemoveClansNotIn
func (u Users) RemoveClansNotIn(tags []string) error {
	return nil
}

// SetClanIn implements db.UsersStore.SetClanIn
func (u Users) SetClanIn(tag string, steamIds []uint64) error {
	return nil
}
