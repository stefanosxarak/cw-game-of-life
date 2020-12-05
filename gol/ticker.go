package gol

import (
	"time"
	"uk.ac.bris.cs/gameoflife/util"
)

type TickerChans struct {
	turnChan    chan int
	done 		chan bool

}

// ticker function
func(t *TickerChans) ticker(c distributorChannels, aliveCells []util.Cell, done <-chan bool) {
	ticker := time.NewTicker(2 * time.Second)
	turn := 0
	running := true
	for running {
		select {
		case <-ticker.C:
			c.events <- AliveCellsCount{turn, len(aliveCells)}
		case <-t.done:
			ticker.Stop()
			running = false
		case currentTurns := <-t.turnChan:
			turn = currentTurns + 1
		default:
			break
		}
	}
}
