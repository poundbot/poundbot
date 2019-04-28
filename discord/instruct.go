package discord

import (
	"bytes"
	"fmt"
	"github.com/poundbot/poundbot/messages"
	"github.com/poundbot/poundbot/types"
	uuid "github.com/satori/go.uuid"
	"log"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
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

type InstructAccountUpdater interface {
	AddServer(snowflake string, server types.Server) error
	UpdateServer(snowflake, oldKey string, server types.Server) error
	RemoveServer(snowflake, serverKey string) error
}

func instruct(botID, channelID, authorID, message string, account types.Account, au InstructAccountUpdater) instructResponse {
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
		log.Println("Instruct without any actual instruction")
		return instructResponse{responseType: instructResponseNone}
	}

	command := parts[0]
	parts = parts[1:]

	if command == "help" {
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
		log.Println("Message is not from owner")
		return instructResponse{responseType: instructResponseNone}
	}

	switch command {
	case "server":
		return instructServer(parts, channelID, guildID, account, au)
	}

	log.Printf("Invalid command %s", command)
	return instructResponse{
		responseType: instructResponseChannel,
		message:      fmt.Sprintf("Invalid command %s.", command),
	}
}

func instructServer(parts []string, channelID, guildID string, account types.Account, au InstructAccountUpdater) instructResponse {
	if len(parts) == 0 {
		return instructResponse{message: "TODO: Server Usage. See `help`."}
	}

	switch parts[0] {
	case "list":
		buf := new(bytes.Buffer)
		w := tabwriter.NewWriter(buf, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "`ID\tName\tRaid Delay\tKey`\t")
		for i, server := range account.Servers {
			fmt.Fprintf(w, "`%d\t%s\t%s\t|`||`%s`||\t\n", i+1, server.Name, server.RaidDelay, server.Key)
		}
		w.Flush()
		out := buf.String()
		return instructResponse{message: out}
	case "add":
		if len(parts) < 2 {
			return instructResponse{
				responseType: instructResponseChannel,
				message:      "Usage: `server add <name>`",
			}
		}
		server := types.Server{
			Name:       strings.Join(parts[1:], " "),
			Key:        uuid.NewV4().String(),
			ChatChanID: channelID,
			RaidDelay:  "1m",
		}
		au.AddServer(guildID, server)
		return instructResponse{message: messages.ServerKeyMessage(server.Name, server.Key)}
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
			return instructResponse{
				responseType: instructResponseChannel,
				message:      "Could not find server command. See `help`.",
			}
		}

		id, err := strconv.Atoi(instructions[0])
		if err != nil {
			return instructResponse{
				responseType: instructResponseChannel,
				message:      "Server ID was not a number. See `server list`",
			}
		}

		if id < 1 {
			return instructResponse{
				responseType: instructResponseChannel,
				message:      "Server ID was not a positive number. See `server list`",
			}
		}

		serverID = id - 1
		instructions = instructions[1:]
	} else if len(account.Servers) > 1 {
		return instructResponse{
			responseType: instructResponseChannel,
			message:      "You must supply the server ID (number). See `server list` or `help",
		}
	}

	if len(account.Servers) < serverID {
		return instructResponse{
			responseType: instructResponseChannel,
			message:      "Server not defined. Try `server list` or `help",
		}
	}

	server := account.Servers[serverID]

	switch instructions[0] {
	case "reset":
		oldKey := server.Key
		server.Key = uuid.NewV4().String()
		au.UpdateServer(guildID, oldKey, server)
		return instructResponse{message: messages.ServerKeyMessage(server.Name, server.Key)}
	case "rename":
		if len(instructions) < 2 {
			return instructResponse{
				responseType: instructResponseChannel,
				message:      "Usage: `server rename [id] <name>`",
			}
		}
		server.Name = strings.Join(instructions[1:], " ")
		au.UpdateServer(guildID, server.Key, server)
		return instructResponse{
			responseType: instructResponseChannel,
			message:      fmt.Sprintf("Server %d name set to %s", serverID+1, server.Name),
		}
	case "delete":
		if err := au.RemoveServer(guildID, server.Key); err != nil {
			return instructResponse{
				responseType: instructResponseChannel,
				message:      "Error removing server. Please try again.",
			}
		}
		return instructResponse{
			responseType: instructResponseChannel,
			message:      fmt.Sprintf("Server %s (%d) removed", server.Name, serverID+1),
		}
	case "chathere":
		server.ChatChanID = channelID
		au.UpdateServer(guildID, server.Key, server)
		return instructResponse{
			responseType: instructResponseChannel,
			message:      fmt.Sprintf("Server %d (%s) will chat here.", serverID+1, server.Name),
		}
	case "raiddelay":
		if len(instructions) != 2 {
			return instructResponse{
				responseType: instructResponseChannel,
				message:      "Usage: `server [ID] rename <name>`",
			}
		}
		_, err := time.ParseDuration(instructions[1])
		if err != nil {
			return instructResponse{
				responseType: instructResponseChannel,
				message:      "Invalid duration format. Examples:\n`1m` = 1 minute, `1h` = 1 hour, `1s` = 1 second",
			}
		}

		server.RaidDelay = instructions[1]
		au.UpdateServer(guildID, server.Key, server)
	}
	return instructResponse{responseType: instructResponseNone}
}
