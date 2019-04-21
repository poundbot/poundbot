package rustconn

import (
	"log"

	"github.com/poundbot/poundbot/storage"
	"github.com/poundbot/poundbot/types"
)

const asLogSymbol = "[AUTH] "

type discordAuthRemover interface {
	Remove(storage.UserInfoGetter) error
}

type userIDGetter interface {
	GetPlayerID() string
	GetDiscordID() string
}

type userUpserter interface {
	UpsertPlayer(storage.UserInfoGetter) error
}

// An AuthSaver saves Discord -> Rust user authentications
type AuthSaver struct {
	das         discordAuthRemover
	us          userUpserter
	authSuccess chan types.DiscordAuth
	done        chan struct{}
}

// NewAuthSaver creates a new AuthSaver
func NewAuthSaver(da discordAuthRemover, u userUpserter, as chan types.DiscordAuth, done chan struct{}) *AuthSaver {
	return &AuthSaver{
		das:         da,
		us:          u,
		authSuccess: as,
		done:        done,
	}
}

// Run updates users sent in through the AuthSuccess channel
func (a *AuthSaver) Run() {
	defer log.Println(asLogSymbol + "AuthServer Stopped.")
	log.Println(asLogSymbol + "Starting AuthServer")
	for {
		select {
		case as := <-a.authSuccess:
			err := a.us.UpsertPlayer(as)

			if err == nil {
				a.das.Remove(as)
				if as.Ack != nil {
					as.Ack(true)
				}
			} else {
				if as.Ack != nil {
					as.Ack(false)
				}
			}
		case <-a.done:
			log.Println(asLogSymbol + "Shutting down AuthServer...")
			return
		}
	}
}
