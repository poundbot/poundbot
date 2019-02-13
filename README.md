# PoundBot

PoundBot is a [Discord](https://discord.gg/) bot for the Rust game server, prividing the following:

* Raid alerts
* Bidirectional chat to a [Discord](https://discord.gg/) channel
  * Requires the installation of [BetterChat](https://umod.org/plugins/better-chat)

## Setup
1. Download PoundBotConnector.cs and add it to your Oxide plugins.
2. Add the bot to your Discord server at http://addpoundbot.mrpoundsign.com/
which 
3. \@mention BoundBot with ``server init`` in the channel you want your chat relay to occur. (Example: ```@PoundBot server init```)
4. Poundbot will whisper you your API key and instructions on where to put it.
   Note that chat will be relayed into the channel you sent this initialization.

### If you want to change the chat relay channel
1. Decide which channel you want the chat relay in.
2. \@mention PoundBot with ``server chat here``

## Raid Alerts
1. In Rust, type /discord "YourUsername#7263"
2. PoundBot should message you asking for the PIN number displayed in chat.
3. Respond to PoundBot with that PIN number and you should be connected!

## Things I can do on the back-end
Add a delay to raid alerts. Currently players will get notified until 1 minute after one of their building structures or other deployed items gets destroyed. I can adjust this if needed.

Anything else you may be having a problem with, I can try to help resolve.

If you have any questions, you can find us on Discord at https://discord.gg/jT3HSUj in the #poundbot_support channel.

Thanks!