package rust

import (
	"fmt"
	"log"
	"time"
)

const wLogSymbol = "📶 "

type Watcher struct {
	querier    Querier
	statusChan chan string
	deltaFreq  int
	done       chan struct{}
}

func NewWatcher(config ServerConfig, pDeltaFreq int, statusChan chan string) (*Watcher, error) {
	rq, err := NewQuerier(config)
	if err != nil {
		return nil, err
	}
	err = rq.Update()
	if err != nil {
		log.Println(wLogSymbol + "⚠️ Error contacting Rust server: " + err.Error())
	}
	return &Watcher{
		querier:    *rq,
		statusChan: statusChan,
		deltaFreq:  pDeltaFreq,
	}, nil
}

func (w *Watcher) Start() error {
	w.done = make(chan struct{})
	log.Println(wLogSymbol + "🛫 Starting Rust Watcher")

	go func() {

		var lastCheck = time.Now().UTC()

		var serverDown = true
		var downChecks uint
		var playerDelta int8

		var lowestPlayers uint8

		var waitOrKill = func(t time.Duration) (kill bool) {
			select {
			case <-w.done:
				log.Println(wLogSymbol + "🛑 Shutting down RustWatcher")
				kill = true
			case <-time.After(t):
				kill = false
			}
			return
		}

		for {
			err := w.querier.Update()
			if err != nil {
				playerDelta = 0
				serverDown = true
				downChecks++
				if downChecks%3 == 0 {
					log.Printf(wLogSymbol+" 🏃 ⚠️ Server is down! %s", err)
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
					log.Println(wLogSymbol + " 🏃 Server is back!")
				}

				if serverDown {
					lastCheck = time.Now().UTC()
					playerDelta = 0
					lowestPlayers = w.querier.PlayerInfo.Players
				}
				serverDown = false
				playerDelta += w.querier.PlayerInfo.PlayersDelta
				if playerDelta < 0 && w.querier.PlayerInfo.Players < lowestPlayers {
					playerDelta = 0
					lowestPlayers = w.querier.PlayerInfo.Players
				}
				// lastUp = time.Now().UTC()
				var now = time.Now().UTC()
				var duration = int(now.Sub(lastCheck).Minutes())
				if playerDelta > 3 || duration >= w.deltaFreq {
					lastCheck = time.Now().UTC()
					if playerDelta > 0 {
						lowestPlayers = w.querier.PlayerInfo.Players
						var playerString = "player has"
						if playerDelta > 1 {
							playerString = "players have"
						}
						message := fmt.Sprintf(
							"@here %d new %s connected, %d of %d playing now!",
							playerDelta,
							playerString,
							w.querier.PlayerInfo.Players,
							w.querier.PlayerInfo.MaxPlayers,
						)
						log.Printf(wLogSymbol+" 🏃 Sending notice of %d new players\n", playerDelta)
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

	return nil
}

func (w *Watcher) Stop() {
	w.done <- struct{}{}
}
