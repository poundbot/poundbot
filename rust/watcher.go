package rust

import (
	"fmt"
	"log"
	"time"
)

type Watcher struct {
	server     Server
	statusChan chan string
	deltaFreq  int
	done       chan struct{}
}

func NewWatcher(config ServerConfig, pDeltaFreq int, statusChan chan string) (*Watcher, error) {
	rs, err := NewServer(config)
	if err != nil {
		return nil, err
	}
	err = rs.Update()
	if err != nil {
		log.Println(logSymbol + " ⚠️ Error contacting Rust server: " + err.Error())
	}
	return &Watcher{
		server:     *rs,
		statusChan: statusChan,
		deltaFreq:  pDeltaFreq,
	}, nil
}

func (w *Watcher) Start() {
	w.done = make(chan struct{})
	log.Println(logSymbol + " Starting Rust Watcher")

	go func() {

		var lastCheck = time.Now().UTC()

		var serverDown = true
		var downChecks uint
		var playerDelta int8

		var lowestPlayers uint8

		var waitOrKill = func(t time.Duration) (kill bool) {
			select {
			case <-w.done:
				log.Println(logSymbol + " Shutting down Rust Watcher")
				kill = true
			case <-time.After(t):
				kill = false
			}
			return
		}

		for {
			err := w.server.Update()
			if err != nil {
				playerDelta = 0
				serverDown = true
				downChecks++
				if downChecks%3 == 0 {
					log.Println(logSymbol + " 🏃 ⚠️ Server is down!")
					if waitOrKill(20 * time.Second) {
						return
					}
				}
				if waitOrKill(5 * time.Second) {
					return
				}
			} else {
				if downChecks > 0 {
					downChecks = 0
					log.Println(logSymbol + " 🏃 Server is back!")
				}

				if serverDown {
					lastCheck = time.Now().UTC()
					playerDelta = 0
					lowestPlayers = w.server.PlayerInfo.Players
				}
				serverDown = false
				playerDelta += w.server.PlayerInfo.PlayersDelta
				if playerDelta < 0 && w.server.PlayerInfo.Players < lowestPlayers {
					playerDelta = 0
					lowestPlayers = w.server.PlayerInfo.Players
				}
				// lastUp = time.Now().UTC()
				var now = time.Now().UTC()
				var duration = int(now.Sub(lastCheck).Minutes())
				if playerDelta > 3 || duration >= w.deltaFreq {
					lastCheck = time.Now().UTC()
					if playerDelta > 0 {
						lowestPlayers = w.server.PlayerInfo.Players
						var playerString = "player has"
						if playerDelta > 1 {
							playerString = "players have"
						}
						message := fmt.Sprintf(
							"@here %d new %s connected, %d of %d playing now!",
							playerDelta,
							playerString,
							w.server.PlayerInfo.Players,
							w.server.PlayerInfo.MaxPlayers,
						)
						log.Printf(logSymbol+" 🏃 Sending notice of %d new players\n", playerDelta)
						w.statusChan <- message
						playerDelta = 0
					}
				}
			}

			if waitOrKill(30 * time.Second) {
				return
			}
		}
	}()
}

func (w *Watcher) Stop() {
	w.done <- struct{}{}
}
