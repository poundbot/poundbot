package jsonstore

import (
	"bitbucket.org/mrpoundsign/poundbot/types"
)

// A Clans implements db.ClansStore
type Clans struct{}

// Upsert does nothing
func (c Clans) Upsert(cl types.Clan) error {
	return nil
}

// Remove does nothing
func (c Clans) Remove(tag string) error {
	return nil
}

// RemoveNotIn does nothing
func (c Clans) RemoveNotIn(tags []string) error {
	return nil
}
