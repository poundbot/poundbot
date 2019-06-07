package types

import "fmt"

type RoleSet struct {
	GuildID   string `json:"-"`
	Role      string
	PlayerIDs []string
}

func (gs *RoleSet) SetGame(game string) {
	for i := range gs.PlayerIDs {
		gs.PlayerIDs[i] = fmt.Sprintf("%s:%s", game, gs.PlayerIDs[i])
	}
}
