# PoundBot Self Hosting

Please note you **WILL NOT** get support for this, but some people have asked for it, so here it is.

## Requirements

* [go 1.11+](https://golang.org)
* [MongoDB](https://mongodb.org]
* a rust server, and PoundbotConnector.cs plugin.

## Running

```go run cmd/poundbot/poundbot.go```

You can also build poundbot and run it. This is outside of the scope of this codument.

## Configuration

### Initialize a new config.json

Create a new configuration file with
```poundbot -init```

### Sample Config

```json
{
  "discord": {
    "token": "YOUR DISCORD BOT AUTH TOKEN"
  },
  "http": {
    "bind_addr": "",
    "port": 9090
  },
  "mongo": {
    "database": "poundbot",
    "dial-addr": "mongodb://localhost"
  }
}
```