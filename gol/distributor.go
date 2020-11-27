package gol

import (
	"fmt"
)

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	iofilename chan<- string
	iooutput   chan<- uint8
	ioinput    <-chan uint8
}

type cell struct {
	x, y int
}

const alive = 255
const dead = 0

func mod(x, m int) int {
	return (x + m) % m
}

func calculateNeighbours(p Params, x, y int, world [][]byte) int {
	neighbours := 0
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			if i != 0 || j != 0 {
				if world[mod(y+i, p.ImageHeight)][mod(x+j, p.ImageWidth)] == alive {
					neighbours++
				}
			}
		}
	}
	return neighbours
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {

	// TODO: Create a 2D slice to store the world.
	world := make([][]byte, p.ImageHeight)
	for i := range world {
		world[i] = make([]byte, p.ImageWidth)
	}
	aliveCells := []cell{}

	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			if world[y][x] == alive {
				aliveCells = append(aliveCells, cell{x: x, y: y})
			}
		}
	}
	// TODO: For all initially alive cells send a CellFlipped Event.
	turn := 0
	c.events <- CellFlipped{}
	c.events <- AliveCellsCount{turn, len(aliveCells)}

	// TODO: Execute all turns of the Game of Life.
	for t := 0; t < p.Turns; t++ {
		for y := 0; y < p.ImageHeight; y++ {
			for x := 0; x < p.ImageWidth; x++ {
				neighbours := calculateNeighbours(p, x, y, world)
				if world[y][x] == alive {
					if neighbours == 2 || neighbours == 3 {
						world[y][x] = alive
					} else {
						world[y][x] = dead
					}
				} else {
					if neighbours == 3 {
						world[y][x] = alive
					} else {
						world[y][x] = dead
					}
				}
			}
		}
		//alivecellscount
		//TurnComplete
	}
	//FinalTurnComplete

	// TODO: Send correct Events when required, e.g. CellFlipped, TurnComplete and FinalTurnComplete.
	//		 See event.go for a list of all events.

	// Make sure that the Io has finished any output before exiting.

	filename := fmt.Sprintf("%dx%d", p.ImageHeight, p.ImageWidth)
	c.iofilename <- filename

	//TODO na ftiaksoume to input kai to output
	// input := make([]byte, p.ImageWidth)
	// output := make([]byte, p.ImageWidth)
	// input <- c.ioinput
	// c.iooutput <- io
	// ImageOutputComplete

	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{turn, Quitting}
	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}
