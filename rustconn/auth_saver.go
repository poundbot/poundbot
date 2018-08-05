package rustconn

import (
	"log"

	"bitbucket.org/mrpoundsign/poundbot/db"
	"bitbucket.org/mrpoundsign/poundbot/types"
)

// An AuthSaver saves Discord -> Rust user authentications
type AuthSaver struct {
	DiscordAuths db.DiscordAuthsStore
	Users        db.UsersStore
	AuthSuccess  chan types.DiscordAuth
	done         chan struct{}
}

// NewAuthSaver creates a new AuthSaver
func NewAuthSaver(da db.DiscordAuthsStore, u db.UsersStore, as chan types.DiscordAuth, done chan struct{}) *AuthSaver {
	return &AuthSaver{
		DiscordAuths: da,
		Users:        u,
		AuthSuccess:  as,
		done:         done,
	}
}

// Run writes users sent in through the AuthSuccess channel
func (a *AuthSaver) Run() {
	for {
		select {
		case as := <-a.AuthSuccess:

			err := a.Users.UpsertBase(as.BaseUser)

			if err == nil {
				a.DiscordAuths.Remove(as.SteamInfo)
				if as.Ack != nil {
					as.Ack(true)
				}
			} else {
				if as.Ack != nil {
					as.Ack(false)
				}
			}
		case <-a.done:
			log.Println(logSymbol + "AuthServer shutting down")
			return
		}
	}
}
