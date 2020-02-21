# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [UNRELEASED]

### Changed

- Updated all module dependencies
- `mongodb.dial-addr` is now `mongodb.dial`
- Hopefully creates less DB connections.
- Env vars can be used for config values.
  - `.` is replaced with `_`
  - examples:
    - `DISCORD_TOKEN=token` sets `discord.token` to `token`
    - Booleans must be 1 or `true` otherwise they are false.
      - `MONGO_SSL_ENABLED=1` sets `mongo.ssl.enabled` to `true`
      - `MONGO_SSL_ENABLED=0` sets `mongo.ssl.enabled` to `false`
      - `MONGO_SSL_ENABLED=true` sets `mongo.ssl.enabled` to `true`
      - `MONGO_SSL_ENABLED=yes` sets `mongo.ssl.enabled` to `false`

### Added

- New DM commands `help`, `status`, `unregister`
  - `help` - displays DM help
  - `status` - displays a users status (registere games)
  - `unregister` - allows a user to remove themselves from poundbot or specific games.

## 4.0.1

### Fixes

- Attempt to fix removal of accounts on false `GuildDelete` from discord/discordgo

## 4.0.0

### Added

- Refactored gameapi handler methods and logging
- Removed deprecated support for discord auth with uint64 SteamID
- Removed version checks in gameapi handlers
- Refactored API between gameapi and discord packages
- New player check API (`GET /discord_auth/check/{player_id}`)
- New roles API (`PUT /roles/{role_name}`)
- New server channels list (`GET /api/messages`)
- Added `CHANGELOG.md`. Hello!

### Changed

- `rustconn` renamed to `gameapi` and other refactors.
- Refactoring in `discord` package to make some methods more testable.
- Refactored `discord.Client` to `discord.Runner`

### Removed

- Removed chat post support in favor of messages API

### Fixes

- Messages sent to embed now properly return an error if permission to post is not available.
