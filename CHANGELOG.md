# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- New roles API (`PUT /roles/{role_name}`)
- New server channels list (`GET /api/messages`)
- Added `CHANGELOG.md`. Hello!

### Changed

- `rustconn` renamed to `gameapi` and several refactorings.
- Refactoring in `discord` package to make some methods more testable.
- Refactored `discord.Client` to `discord.Runner`

### Removed

- Removed chat post support in favor of messages API

### Fixes

- Messages sent to embed now properly return an error if permission to post is not available.
