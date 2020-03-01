# PoundBot

[PoundBot](https://github.com/poundbot/poundbot) is a [Discord](https://discord.gg/) bot for game servers.

Build Status: [![CircleCI](https://circleci.com/gh/poundbot/poundbot.svg?style=svg)](https://circleci.com/gh/poundbot/poundbot)

## Supported games

You can find the plugins for games supported by uMod on [the uMod site](https://umod.org/plugins/pound-bot).

## PoundBot Self Hosting

Please note you **WILL NOT** get support for this, but some people have asked for it, so here it is.

### Requirements

* [go 1.13+](https://golang.org)
* [MongoDB 4.2+](https://mongodb.org)

### Running

`go run cmd/poundbot/poundbot.go`

You can also build poundbot and run it. This is outside of the scope of this document.

### Configuration

#### Initialize a new config.json

Create a new configuration file with
`poundbot -init`

#### Sample Config

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
    "dial": "mongodb://localhost",
    "database": "poundbot",
  },
  "profiler": {
    "port": 6061
  }
}
```

## Getting Help

Find us on Discord on the [PoundBot server](https://discord.gg/ZPNtWEf).

## Thanks

* Talha for Turkish translation and testing!
* Akay from [Game Team War](http://gameteamwar.com/) for the PoundBot logo design.

## License

```text
MIT License

Copyright (c) 2018 MrPoundsign

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
