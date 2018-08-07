package rustconn

import (
	"log"

	"bitbucket.org/mrpoundsign/poundbot/types"
)

const asLogSymbol = "🆔 "

type DiscordAuthsStore interface {
	Remove(types.SteamInfo) error
}

type UsersStore interface {
	UpsertBase(types.BaseUser) error
}

// An AuthSaver saves Discord -> Rust user authentications
type AuthSaver struct {
	DiscordAuths DiscordAuthsStore
	Users        UsersStore
	AuthSuccess  chan types.DiscordAuth
	done         chan struct{}
}

// NewAuthSaver creates a new AuthSaver
func NewAuthSaver(da DiscordAuthsStore, u UsersStore, as chan types.DiscordAuth, done chan struct{}) *AuthSaver {
	return &AuthSaver{
		DiscordAuths: da,
		Users:        u,
		AuthSuccess:  as,
		done:         done,
	}
}

// Run writes users sent in through the AuthSuccess channel
func (a *AuthSaver) Run() {
	log.Println(asLogSymbol + "Starting AuthServer")
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
			log.Println(asLogSymbol + "Shutting Down AuthServer...")
			return
		}
	}
}
