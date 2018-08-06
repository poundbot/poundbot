package jsonstore

import (
	"fmt"

	"bitbucket.org/mrpoundsign/poundbot/types"
)

// A DiscordAuths implements db.DiscordAuthsStore
type DiscordAuths struct{}

// Remove implements db.DiscordAuthsStore.Remove
func (d DiscordAuths) Remove(si types.SteamInfo) error {
	discordAuths.mux.Lock()
	defer discordAuths.mux.Unlock()
	if err := discordAuths.s.driver.Delete(
		discordAuths.s.collection,
		fmt.Sprintf("%d", si.SteamID)); err != nil {
		panic(err)
	}
	delete(discordAuths.d, si.SteamID)
	return nil
}

// Upsert implements db.DiscordAuthsStore.Upsert
func (d DiscordAuths) Upsert(da types.DiscordAuth) error {
	discordAuths.mux.Lock()
	defer discordAuths.mux.Unlock()
	if err := discordAuths.s.driver.Write(
		discordAuths.s.collection,
		fmt.Sprintf("%d", da.SteamID),
		da); err != nil {
		panic(err)
	}
	discordAuths.d[da.SteamID] = da
	return nil
}
