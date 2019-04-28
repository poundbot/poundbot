package discord

import (
	"fmt"
	"github.com/poundbot/poundbot/messages"
	"github.com/poundbot/poundbot/types"
	uuid "github.com/satori/go.uuid"
	"log"
	"strconv"
	"strings"
	"time"
)

type InstructAccountUpdater interface {
	AddServer(snowflake string, server types.Server) error
	UpdateServer(snowflake, oldKey string, server types.Server) error
	RemoveServer(snowflake, serverKey string) error
}

type InstructResponder interface {
	sendChannelMessage(channelID, content string) error
	sendPrivateMessage(authorID, message string) error
	sendServerKey(snowflake, name, key string) error
}

func instruct(botID, channelID, authorID, message string, account types.Account, ir InstructResponder, au InstructAccountUpdater) {
	guildID := account.GuildSnowflake
	adminIDs := account.GetAdminIDs()
	log.Printf(
		"Instruct: ID:\"%s\", Author:\"%s\", Guild:\"%s\", Admnins:\"%s\", Message: \"%s\"",
		account.ID, authorID, guildID, account.GetAdminIDs(), message,
	)
	parts := strings.Fields(
		strings.Replace(
			strings.Replace(message, fmt.Sprintf("<@%s>", botID), "", -1),
			fmt.Sprintf("<@!%s>", botID), "", -1,
		),
	)

	if len(parts) == 0 {
		log.Println("Mention without instruction")
		return
	}

	command := parts[0]
	parts = parts[1:]

	if command == "help" {
		log.Printf("Sending help to %s", authorID)
		ir.sendPrivateMessage(authorID, messages.HelpText())
		return
	}

	isOwner := false

	for i := range adminIDs {
		if authorID == adminIDs[i] {
			isOwner = true
			break
		}
	}

	if !isOwner {
		log.Println("Message is not from owner")
		return
	}

	switch command {
	case "help":
		log.Printf("Sending help to %s", authorID)
		ir.sendPrivateMessage(authorID, messages.HelpText())
		break
	case "server":
		if len(parts) == 0 {
			ir.sendChannelMessage(channelID, "TODO: Server Usage. See `help`.")
			return
		}

		switch parts[0] {
		case "list":
			out := "**Server List**\nID : Name : RaidDelay : Key\n----"
			for i, server := range account.Servers {
				out = fmt.Sprintf("%s\n%d : %s : %s : ||`%s`||", out, i+1, server.Name, server.RaidDelay, server.Key)
			}
			ir.sendPrivateMessage(authorID, out)
			return
		case "add":
			if len(parts) < 2 {
				ir.sendChannelMessage(channelID, "Usage: `server add <name>`")
				return
			}
			server := types.Server{
				Name:       strings.Join(parts[1:], " "),
				Key:        uuid.NewV4().String(),
				ChatChanID: channelID,
				RaidDelay:  "1m",
			}
			au.AddServer(guildID, server)
			ir.sendServerKey(authorID, server.Name, server.Key)
			return
		}

		var commands = []string{"reset", "rename", "delete", "chathere", "raiddelay"}
		isCommand := func(s string) bool {
			for _, command := range commands {
				if s == command {
					return true
				}
			}
			return false
		}

		serverID := 0
		instructions := parts

		if !isCommand(instructions[0]) {
			if len(instructions) < 2 || !isCommand(instructions[1]) {
				ir.sendChannelMessage(channelID, "Could not find server command. See `help`.")
				return
			}

			id, err := strconv.Atoi(instructions[0])
			if err != nil {
				ir.sendChannelMessage(channelID, "Server ID was not a number. See `server list`")
				return
			}
			if id < 1 {
				ir.sendChannelMessage(channelID, "Server ID was not a positive number. See `server list`")
				return
			}
			serverID = id - 1
			instructions = instructions[1:]
		} else if len(account.Servers) > 1 {
			ir.sendChannelMessage(channelID, "You must supply the server ID (number). See `server list` or `help")
			return
		}

		if len(account.Servers) < serverID {
			ir.sendChannelMessage(channelID, "Server not defined. Try `server list` or `help")
			return
		}

		server := account.Servers[serverID]

		switch instructions[0] {
		case "reset":
			oldKey := server.Key
			server.Key = uuid.NewV4().String()
			au.UpdateServer(guildID, oldKey, server)
			ir.sendServerKey(authorID, server.Name, server.Key)
			return
		case "rename":
			if len(instructions) < 2 {
				ir.sendChannelMessage(channelID, "Usage: `server rename [id] <name>`")
				return
			}
			server.Name = strings.Join(instructions[1:], " ")
			au.UpdateServer(guildID, server.Key, server)
			ir.sendChannelMessage(channelID, fmt.Sprintf("Server %d name set to %s", serverID+1, server.Name))
			return
		case "delete":
			if err := au.RemoveServer(guildID, server.Key); err != nil {
				ir.sendChannelMessage(channelID, "Error removing server. Please try again.")
			}
			ir.sendChannelMessage(channelID, fmt.Sprintf("Server %s (%d) removed", server.Name, serverID+1))
			return
		case "chathere":
			server.ChatChanID = channelID
			au.UpdateServer(guildID, server.Key, server)
			return
		case "raiddelay":
			if len(instructions) != 2 {
				ir.sendChannelMessage(channelID, "Usage: `server [ID] rename <name>`")
				return
			}
			_, err := time.ParseDuration(instructions[1])
			if err != nil {
				ir.sendChannelMessage(channelID, "Invalid duration format. Examples:\n`1m` = 1 minute, `1h` = 1 hour, `1s` = 1 second")
				return
			}

			server.RaidDelay = instructions[1]
			au.UpdateServer(guildID, server.Key, server)

			return
		}

		log.Printf("Invalid command %s", command)
		ir.sendChannelMessage(
			channelID,
			fmt.Sprintf("Invalid command %s. Are you using the ID from `server list`?", instructions[0]),
		)
	}
}
