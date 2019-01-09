package time

import "github.com/benbjohnson/clock"

var mocking = false
var currentClock = clock.New()

func Clock() clock.Clock {
	return currentClock
}

func Mock() {
	currentClock = clock.NewMock()
}
