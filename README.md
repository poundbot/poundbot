[PoundBot](https://github.com/poundbot/poundbot) is a [Discord](https://discord.gg/) bot for game servers.

This plugin is the base communication layer to the bot. You must intall other plugins to provide additional
functionality. Currently, the bot supports the following:

* [Pound Bot Chat Relay](https://umod.org/plugins/pound-bot-chat-relay) - Bidirectional chat to a [Discord](https://discord.gg/) channel (Universal)
* [Pound Bot Clans](https://umod.org/plugins/pound-bot-clans) - Clan syncing (Universal)
* [Pound Bot Raid Alerts](https://umod.org/plugins/pound-bot-raid-alerts) - Raid alerts for Rust

## Setup

** Important update! As of 1.1.1, we now have a HTTPS API. Please set your `api_url` to `https://api.poundbot.com/` **

1. Download `PoundBot.cs` and add it to your plugins directory.
2. Add the bot to your Discord server at [https://add.poundbot.com/](https://add.poundbot.com/) 
3. Command PoundBot with `!pb server add myservername` in the channel you want your chat relay to occur.
4. PoundBot will whisper you your API key and instructions on where to put it.

## Configuration

### Default Configuration

```json
{
  "api_url": "https://api.poundbot.com/",
  "api_key": "API KEY HERE"
}
```

## Authenticating with PoundBot

Authenticating with PoundBot associates your stem account with discord. This is necessary to determine
who to send messages to from their in-game identity.

1. In chat, type `/pbreg "Your Username#7263"`
2. PoundBot should message you asking for the PIN number displayed in chat.
3. Respond to PoundBot with that PIN number and you should be connected!

## Getting Help

Find us on Discord on the [PoundBot server](https://discord.gg/ZPNtWEf).

## Thanks
Talha for Turkish translation and testing!