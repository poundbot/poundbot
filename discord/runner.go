package discord

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/poundbot/poundbot/pbclock"
	"github.com/poundbot/poundbot/storage"
	"github.com/poundbot/poundbot/types"

	"github.com/bwmarrin/discordgo"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/sirupsen/logrus"
)

var iclock = pbclock.Clock

type Runner struct {
	session         *discordgo.Session
	cqs             storage.ChatQueueStore
	as              storage.AccountsStore
	mls             storage.MessageLocksStore
	das             storage.DiscordAuthsStore
	us              storage.UsersStore
	token           string
	status          chan bool
	chatChan        chan types.ChatMessage
	raidAlertChan   chan types.RaidAlert
	gameMessageChan chan types.GameMessage
	authChan        chan types.DiscordAuth
	AuthSuccess     chan types.DiscordAuth
	channelsRequest chan types.ServerChannelsRequest
	roleSetChan     chan types.RoleSet
	shutdown        bool
}

func NewRunner(token string, as storage.AccountsStore, das storage.DiscordAuthsStore,
	us storage.UsersStore, mls storage.MessageLocksStore, cqs storage.ChatQueueStore) *Runner {
	return &Runner{
		cqs:             cqs,
		mls:             mls,
		as:              as,
		das:             das,
		us:              us,
		token:           token,
		chatChan:        make(chan types.ChatMessage),
		authChan:        make(chan types.DiscordAuth),
		AuthSuccess:     make(chan types.DiscordAuth),
		raidAlertChan:   make(chan types.RaidAlert),
		gameMessageChan: make(chan types.GameMessage),
		channelsRequest: make(chan types.ServerChannelsRequest),
		roleSetChan:     make(chan types.RoleSet),
	}
}

// Start starts the runner
func (r *Runner) Start() error {
	session, err := discordgo.New("Bot " + r.token)
	if err == nil {
		r.session = session
		r.session.AddHandler(r.messageCreate)
		r.session.AddHandler(r.ready)
		r.session.AddHandler(disconnected(r.status))
		r.session.AddHandler(r.resumed)
		r.session.AddHandler(newGuildCreate(r.as, r.us))
		r.session.AddHandler(newGuildDelete(r.as))
		r.session.AddHandler(newGuildMemberAdd(r.us, r.as))
		r.session.AddHandler(newGuildMemberRemove(r.us, r.as))

		r.status = make(chan bool)

		go r.runner()

		connect(session)
	}
	return err
}

func (r Runner) RaidNotify(ra types.RaidAlert) {
	r.raidAlertChan <- ra
}

func (r Runner) AuthDiscord(da types.DiscordAuth) {
	r.authChan <- da
}

func (r Runner) SendChatMessage(cm types.ChatMessage) {
	r.chatChan <- cm
}

// SendGameMessage sends a message from the game to a discord channel
func (r Runner) SendGameMessage(gm types.GameMessage, timeout time.Duration) error {
	select {
	case r.gameMessageChan <- gm:
		return nil
	case <-time.After(timeout):
		return errors.New("no response from discord handler")
	}
}

// ServerChannels sends a request to get the visible chnnels for a discord guild
func (r Runner) ServerChannels(scr types.ServerChannelsRequest) {
	r.channelsRequest <- scr
}

func (r Runner) SetRole(rs types.RoleSet, timeout time.Duration) error {
	// sending message
	select {
	case r.roleSetChan <- rs:
		return nil
	case <-time.After(timeout):
		return errors.New("no response from discord handler")
	}
}

// Stop stops the runner
func (r *Runner) Stop() {
	defer r.session.Close()
	log.WithFields(logrus.Fields{"sys": "RUNNER"}).Info(
		"Disconnecting...",
	)

	r.shutdown = true
}

func (r *Runner) runner() {
	rLog := log.WithFields(logrus.Fields{"sys": "RUNNER"})
	defer rLog.Warn("Runner exited")

	connectedState := false

	for {
		if connectedState {
			rLog.Info("Waiting for messages.")
		Reading:
			for {
				select {
				case connectedState = <-r.status:
					if !connectedState {
						rLog.Warn("Received disconnected message")
						if r.shutdown {
							return
						}
						break Reading
					}

					rLog.Info("Received unexpected connected message")
				case raidAlert := <-r.raidAlertChan:
					raLog := rLog.WithFields(logrus.Fields{"chan": "RAID", "pID": raidAlert.PlayerID})
					raLog.Trace("Got raid alert")
					go func() {
						raUser, err := r.us.GetByPlayerID(raidAlert.PlayerID)
						if err != nil {
							raLog.WithError(err).Error("Player not found trying to send raid alert")
							return
						}

						user, err := r.session.User(raUser.Snowflake)
						if err != nil {
							raLog.WithField("uID", raUser.Snowflake).WithError(err).Error(
								"Discord user not found trying to send raid alert",
							)
							return
						}

						id, err := r.sendPrivateMessage(user.ID, raidAlert.String())
						if err != nil {
							raLog.WithError(err).Error("could not create private channel to send to user")
							return
						}

						raidAlert.MessageIDChannel <- id
						close(raidAlert.MessageIDChannel)
					}()
				case da := <-r.authChan:
					go r.discordAuthHandler(da)
				case m := <-r.gameMessageChan:
					go gameMessageHandler(r.session.State.User.ID, m, r.session.State.Guild, r)
				case cm := <-r.chatChan:
					go gameChatHandler(r.session.State.User.ID, cm, r.session.State.Guild, r)
				case cr := <-r.channelsRequest:
					go sendChannelList(r.session.State.User.ID, cr.GuildID, cr.ResponseChan, r.session.State)
				case rs := <-r.roleSetChan:
					go rolesSetHandler(r.session.State.User.ID, rs, r.session.State, r.us, r.session)
				}
			}
		}
	Connecting:
		for {
			rLog.Info("Waiting for connected state...")
			connectedState = <-r.status
			if connectedState {
				rLog.WithField("sys", "CONN").Info("Received connected message")
				break Connecting
			}
			rLog.WithField("sys", "CONN").Info("Received disconnected message")
		}
	}

}

func (r *Runner) discordAuthHandler(da types.DiscordAuth) {
	dLog := log.WithFields(logrus.Fields{
		"chan": "DAUTH",
		"gID":  da.GuildSnowflake,
		"name": da.DiscordInfo.DiscordName,
		"uID":  da.Snowflake,
	})
	dLog.Trace("Got discord auth")
	dUser, err := r.getUserByName(da.GuildSnowflake, da.DiscordInfo.DiscordName)
	if err != nil {
		dLog.WithError(err).Error("Discord user not found")
		err = r.das.Remove(da)
		if err != nil {
			dLog.WithError(err).Error("Error removing discord auth for PlayerID from the database.")
		}
		return
	}

	da.Snowflake = dUser.ID

	err = r.das.Upsert(da)
	if err != nil {
		dLog.WithError(err).Error("Error upserting PlayerID ito the database")
		return
	}

	_, err = r.sendPrivateMessage(da.Snowflake,
		localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "UserPINPrompt",
				Other: "Enter the PIN provided in-game to validate your account.\nOnce you are validated, you will begin receiving raid alerts!",
			},
		}),
	)

	if err != nil {
		dLog.WithError(err).Error("Could not send PIN request to user")
	}
}

func (r *Runner) resumed(s *discordgo.Session, event *discordgo.Resumed) {
	log.WithField("sys", "CONN").Info("Resumed connection")
	r.status <- true
}

// This function will be called (due to AddHandler above) when the bot receives
// the "ready" event from Discord.
func (r *Runner) ready(s *discordgo.Session, event *discordgo.Ready) {
	log.WithField("sys", "CONN").Info("Connection Ready")

	if err := s.UpdateStatus(0, localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "DiscordStatus",
			Other: "!pb help",
		}}),
	); err != nil {
		log.WithError(err).Error("failed to update bot status")
	}

	guilds := make([]types.BaseAccount, len(s.State.Guilds))
	for i, guild := range s.State.Guilds {
		guilds[i] = types.BaseAccount{GuildSnowflake: guild.ID, OwnerSnowflake: guild.OwnerID}
	}
	if err := r.as.RemoveNotInDiscordGuildList(guilds); err != nil {
		log.WithError(err).Error("could not sync discord guilds")
	}
	r.status <- true
}

// Returns nil user if they don't exist; Returns error if there was a communications error
func (r *Runner) getUserByName(guildID, name string) (discordgo.User, error) {
	guild, err := r.session.State.Guild(guildID)
	if err != nil {
		return discordgo.User{}, fmt.Errorf("guild %s not found searching for user %s", guildID, name)
	}

	for _, user := range guild.Members {
		if strings.ToLower(user.User.String()) == strings.ToLower(name) {
			return *user.User, nil
		}
	}

	return discordgo.User{}, fmt.Errorf("discord user not found %s", name)
}
