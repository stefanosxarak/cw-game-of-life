package gol

import (
	"time"
)

type TickerChans struct {
	tick chan bool
	done chan bool
}

// ticker function
func (t *TickerChans) ticker(events chan<- Event) {
	ticker := time.NewTicker(2 * time.Second)

	select {
	case <-t.done:
		ticker.Stop()
	case <-ticker.C:
		t.tick <- true

	}

}
