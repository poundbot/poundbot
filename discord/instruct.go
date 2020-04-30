package discord

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/gofrs/uuid"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/poundbot/poundbot/messages"
	"github.com/poundbot/poundbot/types"
	"github.com/sirupsen/logrus"
)

type instructResponseType int

const (
	instructResponsePrivate = iota // 0 Response should be sent privately
	instructResponseChannel        // 1 Response should be sent to channel
	instructResponseNone           // 2 No response
)

type instructResponse struct {
	responseType instructResponseType
	message      string
}

type instructAccountUpdater interface {
	AddServer(snowflake string, server types.AccountServer) error
	UpdateServer(snowflake, oldKey string, server types.AccountServer) error
	RemoveServer(snowflake, serverKey string) error
}

func instruct(botID, channelID, authorID, message string, account types.Account, au instructAccountUpdater) instructResponse {
	guildID := account.GuildSnowflake
	adminIDs := account.GetAdminIDs()
	iLog := log.WithFields(logrus.Fields{
		"sys":      "INST",
		"account":  account.ID.Hex(),
		"authorid": authorID,
		"guildid":  guildID,
		"admins":   adminIDs,
		"message":  message,
	})

	iLog.Trace(localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "Instruct",
			Other: "Instruct",
		},
	}))

	parts := getQuotedParts(
		strings.Replace(
			strings.Replace(message, fmt.Sprintf("<@%s>", botID), "", -1),
			fmt.Sprintf("<@!%s>", botID), "", -1,
		),
	)

	if len(parts) == 0 {
		iLog.Trace("Received instruct with no instructions")
		return instructResponse{responseType: instructResponseNone}
	}

	command := parts[0]
	parts = parts[1:]

	if command == localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "InstructCommandHelp",
			Other: "help",
		},
	}) {
		return instructResponse{message: messages.HelpText()}
	}

	isOwner := false

	for i := range adminIDs {
		if authorID == adminIDs[i] {
			isOwner = true
			break
		}
	}

	if !isOwner {
		iLog.Trace("Instruction is not from an admin")
		return instructResponse{
			responseType: instructResponseChannel,
			message: localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "InstructNotAuthorized",
					Other: "You are not authorized to use this commamd. This is only available to the server owner.",
				},
			}),
		}
	}

	switch command {
	case localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "InstructCommandServer",
			Other: "server",
		},
	}):
		return instructServer(parts, channelID, guildID, account, au)
	}

	msg := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "InvalidCommand",
			Other: "Invalid Command: {{.Command}}",
		},
		TemplateData: map[string]string{"Command": command},
	})
	iLog.Trace(msg)

	return instructResponse{
		responseType: instructResponseChannel,
		message:      msg,
	}
}

func instructServer(parts []string, channelID, guildID string, account types.Account, au instructAccountUpdater) instructResponse {
	isLog := log.WithFields(logrus.Fields{"sys": "instructServer",
		"gID":       guildID,
		"cID":       channelID,
		"accountID": account.ID.Hex(),
	})
	if len(parts) == 0 {
		isLog.Trace("Empty instruct")
		return instructResponse{message: "TODO: Server Usage. See `help`."}
	}

	listCmd := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "InstructCommandServerList",
			Other: "list",
		},
	})

	addCmd := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "InstructCommandServerAdd",
			Other: "add",
		},
	})

	switch parts[0] {
	case listCmd:
		isLog = isLog.WithField("cmd", "server list")
		isLog.Trace("server list")
		buf := new(bytes.Buffer)
		w := tabwriter.NewWriter(buf, 0, 0, 3, ' ', 0)

		fmt.Fprintln(w, localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "InstructCommandServerListHeader",
				Other: "`ID\tName\tRaid Delay\tKey`\t",
			},
		}))
		for i, server := range account.Servers {
			fmt.Fprintf(w, "`%d\t%s\t%s\t|`||`%s`||\t\n", i+1, server.Name, server.RaidDelay, server.Key)
		}
		w.Flush()
		out := buf.String()
		return instructResponse{message: out}
	case addCmd:
		isLog = isLog.WithField("cmd", "server add")
		isLog.Trace("server add")
		if len(parts) < 2 {
			return instructResponse{
				responseType: instructResponseChannel,
				message: localizer.MustLocalize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "InstructCommandServerAddUsage",
						Other: "Usage: `server add <name>`",
					},
				}),
			}
		}
		ruid, err := uuid.NewV4()
		if err != nil {
			isLog.WithError(err).Error("error creating uuid")
			return instructResponse{responseType: instructResponseNone}
		}
		server := types.AccountServer{
			Name:                strings.Join(parts[1:], " "),
			Key:                 ruid.String(),
			Channels:            []types.AccountServerChannel{{ChannelID: channelID, Tags: []string{"chat", "serverchat"}}},
			RaidDelay:           "1m",
			RaidNotifyFrequency: "10m",
		}
		err = au.AddServer(guildID, server)
		if err != nil {
			isLog.WithError(err).Error("could not add server")
			return instructResponse{message: "Internal error adding server. Please try again."}
		}
		return instructResponse{message: messages.ServerKeyMessage(server.Name, server.Key)}
	}

	serverID, instructions, err := instructServerArgs(parts, account.Servers)
	if err != nil {
		message := "Error processing server command. See `help`."
		switch err.Error() {
		case "invalid server id":
			message = localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "InstructResponseInvalidServerID",
					Other: "Invalid server ID. See `help`.",
				},
			})
		case "server id required":
			message = localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "InstructResponseMultipleServersDefines",
					Other: "You have multiple servers defined. You must supply a server ID. See `server list` or `help`.",
				},
			})
		}

		return instructResponse{
			responseType: instructResponseChannel,
			message:      message,
		}
	}

	resetCmd := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "InstructCommandServerReset",
			Other: "reset",
		},
	})

	renameCmd := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "InstructCommandServerRename",
			Other: "rename",
		},
	})

	deleteCmd := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "InstructCommandServerDelete",
			Other: "delete",
		},
	})

	chathereCmd := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "InstructCommandServerChatHere",
			Other: "chathere",
		},
	})

	raidDelayCmd := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "InstructCommandServerRaidDelay",
			Other: "raiddelay",
		},
	})

	raidNotifyFrequencyCmd := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "InstructCommandServerRaidNotifyFrequency",
			Other: "raidnotifyfrequency",
		},
	})

	if len(account.Servers)-1 < serverID {
		return instructResponse{
			responseType: instructResponseChannel,
			message: localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "InstructCommandServerDoesNotExist",
					Other: "Invalid server ID. Check server list.",
				},
			}),
		}
	}
	server := account.Servers[serverID]

	switch instructions[0] {
	case resetCmd:
		isLog = isLog.WithField("cmd", "server reset")
		isLog.Trace("server reset")
		ruid, err := uuid.NewV4()
		if err != nil {
			isLog.WithError(err).Error("error creating uuid")
			return instructResponse{responseType: instructResponseNone}
		}
		oldKey := server.Key
		server.Key = ruid.String()

		if err = au.UpdateServer(guildID, oldKey, server); err != nil {
			isLog.WithError(err).Error("storage error updating server")
		}

		return instructResponse{message: messages.ServerKeyMessage(server.Name, server.Key)}
	case renameCmd:
		isLog = isLog.WithField("cmd", "server rename")
		isLog.Trace("server rename")
		if len(instructions) < 2 {
			return instructResponse{
				responseType: instructResponseChannel,
				message: localizer.MustLocalize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "InstructCommandServerRenameUsage",
						Other: "Usage: `server [id] rename <name>`",
					},
				}),
			}
		}
		server.Name = strings.Join(instructions[1:], " ")

		if err = au.UpdateServer(guildID, server.Key, server); err != nil {
			isLog.WithError(err).Error("storage error updating server")
		}

		return instructResponse{
			responseType: instructResponseChannel,
			message:      fmt.Sprintf("Server %d name set to %s", serverID+1, server.Name),
		}
	case deleteCmd:
		if err := au.RemoveServer(guildID, server.Key); err != nil {
			return instructResponse{
				responseType: instructResponseChannel,
				message:      "Error removing server. Please try again.",
			}
		}
		return instructResponse{
			responseType: instructResponseChannel,
			message: localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "InstructCommandServerDeleteResponse",
					Other: "Server {{.Name}} ({{.ID}}) removed",
				},
				TemplateData: map[string]string{
					"Name": server.Name,
					"ID":   fmt.Sprint(serverID + 1),
				},
			}),
		}
	case chathereCmd:
		isLog = isLog.WithField("cmd", "server chathere")
		isLog.Trace("server chathere")
		server.SetChannelIDForTag(channelID, "chat")
		server.SetChannelIDForTag(channelID, "serverchat")

		if err = au.UpdateServer(guildID, server.Key, server); err != nil {
			isLog.WithError(err).Error("storage error updating server")
			return instructResponse{message: "Internal error. Please try again."}
		}

		return instructResponse{
			responseType: instructResponseChannel,
			message: localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "InstructCommandServerChatHereResponse",
					Other: "Server {{.Name}} ({{.ID}}) will chat here",
				},
				TemplateData: map[string]string{
					"Name": server.Name,
					"ID":   fmt.Sprint(serverID + 1),
				},
			}),
		}
	case raidDelayCmd:
		isLog = isLog.WithField("cmd", "server raidDelay")
		isLog.Trace("server raidDelay")
		if len(instructions) != 2 {
			return instructResponse{
				responseType: instructResponseChannel,
				message: localizer.MustLocalize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "InstructCommandServerRaidDelayUsage",
						Other: "Usage: `server [id] raiddelay <duration>`",
					},
				}),
			}
		}
		_, err := time.ParseDuration(instructions[1])
		if err != nil {
			return instructResponse{
				responseType: instructResponseChannel,
				message: localizer.MustLocalize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "InstructCommandServerRaidDelayInvalidFormat",
						Other: "Invalid duration format. Examples: `5m` = 5 minutes, `1h` = 1 hour, `1s` = 1 second",
					},
				}),
			}
		}

		server.RaidDelay = instructions[1]

		if err = au.UpdateServer(guildID, server.Key, server); err != nil {
			isLog.WithError(err).Error("storage error updating server")
			return instructResponse{message: "Internal error. Please try again."}
		}

		return instructResponse{
			responseType: instructResponseChannel,
			message: localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "InstructCommandServerChatHereResponse",
					Other: "RaidDelay for {{.ID}}:{{.Name}} is now {{.RaidDelay}}",
				},
				TemplateData: map[string]string{
					"Name":      server.Name,
					"ID":        fmt.Sprint(serverID + 1),
					"RaidDelay": server.RaidDelay,
				},
			}), //fmt.Sprintf("RaidDelay for %d:%s is now %s", serverID+1, server.Name, server.RaidDelay),
		}

	case raidNotifyFrequencyCmd:
		isLog = isLog.WithField("cmd", "server raidNotifyFrequency")
		isLog.Trace("server raidNotifyFrequency")
		if len(instructions) != 2 {
			return instructResponse{
				responseType: instructResponseChannel,
				message: localizer.MustLocalize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "InstructCommandServerRaidNotificationFrequencyUsage",
						Other: "Usage: `server [id] raidnotificationfrequency <duration>`",
					},
				}),
			}
		}
		_, err := time.ParseDuration(instructions[1])
		if err != nil {
			return instructResponse{
				responseType: instructResponseChannel,
				message: localizer.MustLocalize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "InstructCommandServerRaidNotificationFrequencyInvalidFormat",
						Other: "Invalid duration format. Examples: `5m` = 5 minutes, `1h` = 1 hour, `1s` = 1 second",
					},
				}),
			}
		}

		server.RaidNotifyFrequency = instructions[1]

		if err = au.UpdateServer(guildID, server.Key, server); err != nil {
			isLog.WithError(err).Error("storage error updating server")
			return instructResponse{message: "Internal error. Please try again."}
		}

		return instructResponse{
			responseType: instructResponseChannel,
			message: localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "InstructCommandServerChatHereResponse",
					Other: "RaidNotifyFrequency for {{.ID}}:{{.Name}} is now {{.RaidNotifyFrequency}}",
				},
				TemplateData: map[string]string{
					"Name":                server.Name,
					"ID":                  fmt.Sprint(serverID + 1),
					"RaidNotifyFrequency": server.RaidDelay,
				},
			}), //fmt.Sprintf("RaidDelay for %d:%s is now %s", serverID+1, server.Name, server.RaidDelay),
		}
	}
	return instructResponse{responseType: instructResponseNone}
}

func instructServerArgs(parts []string, servers []types.AccountServer) (int, []string, error) {
	resetCmd := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "InstructCommandServerReset",
			Other: "reset",
		},
	})

	renameCmd := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "InstructCommandServerRename",
			Other: "rename",
		},
	})

	deleteCmd := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "InstructCommandServerDelete",
			Other: "delete",
		},
	})

	chathereCmd := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "InstructCommandServerChatHere",
			Other: "chathere",
		},
	})

	raidDelayCmd := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "InstructCommandServerRaidDelay",
			Other: "raiddelay",
		},
	})

	var serverID int
	var commands = []string{resetCmd, renameCmd, deleteCmd, chathereCmd, raidDelayCmd}
	isCommand := func(s string) bool {
		for i := range commands {
			if s == commands[i] {
				return true
			}
		}
		return false
	}

	if !isCommand(parts[0]) {
		if len(parts) < 2 || !isCommand(parts[1]) {
			return -1, []string{}, errors.New("invalid command")
		}

		id, err := strconv.Atoi(parts[0])
		if err != nil {
			return -1, []string{}, errors.New("invalid server id")
		}

		if id < 1 {
			return -1, []string{}, errors.New("invalid server id")
		}

		serverID = id - 1
		parts = parts[1:]
	} else if len(servers) > 1 {
		return -1, []string{}, errors.New("server id required")
	}

	if len(servers) < serverID {
		return -1, []string{}, errors.New("invalid server id")
	}

	return serverID, parts, nil
}
