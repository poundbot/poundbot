package messages

import (
	"fmt"
	"strings"
)

const HelpText = `
Commands:
  server init       - Initializes your server and will PM you the API key.
                      Chat will be relayed into the channel you send this
                      message from.
  server reset      - Resets your server API Key. A new key will be sent to you.
  server chat here  - Sets the channel for server chat to this channel
  
  Download the plugin at https://bitbucket.org/mrpoundsign/poundbot/src/multi-server/rust_plugin/
`

const PinPrompt = `
Enter the PIN provided in-game to validate your account.
Once you are validated, you will begin receiving raid alerts!
`

func ServerKeyMessage(key string) string {
	return fmt.Sprintf("Your new server key is *%s*. Add it to your oxide/config/PoundBotConnextor.json or copy and paste the following:\n\n```"+`
{
	"api_url": "http://poundbot.mrpoundsign.com:7070/",
	"show_own_damage": true,
	"api_key": "%s"
}
`+"```", key, key)
}

func RaidAlert(serverName string, gridPositions, items []string) string {
	return fmt.Sprintf(`
%s RAID ALERT! You are being raided!

  Locations:
    %s

  Destroyed:
    %s
`, serverName, strings.Join(gridPositions, ", "), strings.Join(items, ", "))
}
