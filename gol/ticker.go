package gol

import (
	"time"
)

type TickerChans struct {
	tick chan bool
	done chan bool
}

// ticker function
func (t *TickerChans) ticker(done <-chan bool) {
	ticker := time.NewTicker(2 * time.Second)

	select {
	case <-ticker.C:
		t.tick <- true
	case <-t.done:
		ticker.Stop()

	}

}
