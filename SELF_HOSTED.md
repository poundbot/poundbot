## Requirements

* [go 1.11+](https://golang.org)
* [MongoDB](https://mongodb.org]
* a rust server, and PoundbotConnector.cs plugin.

## To run:

```go run poundbot.go```

## Configuration
Config can be in the following locations:

```
/etc/poundbot/config.json
~/.poundbot/config.json
config.json
```

### Sample Config
```
{
    "twitter": {
        "userid": <twitter user id from http://gettwitterid.com/>,
        "filters": [
            "#almupdate"
        ],
        "consumer": {
            "key": "KEY",
            "secret": "SECRET"
        },
        "access": {
            "token": "TOKEN",
            "secret": "SECRET"
        }
    },
    "discord": {
        "channels": {
            "link": "CHANID1 (right click on discord client channel to obtain IDs)",
            "status": "CHANID2",
            "bot": "CHANID3",
            "general": "CHANID4"
        },
        "token": "bot token"
    },
    "rust": {
        "server": {
            "hostname": "rust.example.com",
            "port": 28015
        }
    }
}
```