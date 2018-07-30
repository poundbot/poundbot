package db

import "mrpoundsign.com/poundbot/types"

func RemoveDiscordAuth(s *Session, si types.SteamInfo) {
	s.DiscordAuthCollection().Remove(si)
}

func UpsertDiscordAuth(s *Session, da types.DiscordAuth) error {
	_, err := s.DiscordAuthCollection().Upsert(da.SteamInfo, da)
	return err
}
