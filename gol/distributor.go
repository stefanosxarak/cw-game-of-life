package gol

import (
	"fmt"

	"uk.ac.bris.cs/gameoflife/util"
)

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioOutput   chan<- uint8
	ioInput    <-chan uint8
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

func calculateNextState(p Params, world [][]byte) [][]byte {
	newWorld := make([][]byte, p.ImageHeight)
	for i := range newWorld {
		newWorld[i] = make([]byte, p.ImageWidth)
	}
	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			neighbours := calculateNeighbours(p, x, y, world)
			if world[y][x] == alive {
				if neighbours == 2 || neighbours == 3 {
					newWorld[y][x] = alive
				} else {
					newWorld[y][x] = dead
				}
			} else {
				if neighbours == 3 {
					newWorld[y][x] = alive
				} else {
					newWorld[y][x] = dead
				}
			}
		}
	}
	return newWorld
}

func calculateAliveCells(p Params, world [][]byte) []cell {
	aliveCells := []cell{}

	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			if world[y][x] == alive {
				aliveCells = append(aliveCells, cell{x: x, y: y})
			}
		}
	}

	return aliveCells
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {

	fmt.Println("filename check 1")
	//imageName := fmt.Sprintf("%dx%d", p.ImageHeight, p.ImageWidth)
	//imageName := fmt.Sprintf("images/%vx%v.pgm", p.ImageWidth, p.ImageHeight)
	// imageName := fmt.Sprintf("%vx%v", p.ImageWidth, p.ImageHeight)

	// data, ioError := ioutil.ReadFile("images/" + filename + ".pgm")
	// c.ioFilename <- imageName
	fmt.Println("filename check 2")

	fmt.Println("input check 1")
	c.ioCommand <- ioInput
	<-c.ioInput
	fmt.Println("input check 2")

	// TODO: Create a 2D slice to store the world.
	newworld := make([][]byte, p.ImageHeight)
	for i := range newworld {
		newworld[i] = make([]byte, p.ImageWidth)
	}

	aliveCells := calculateAliveCells(p, newworld)
	// TODO: For all initially alive cells send a CellFlipped Event.
	turn := 0
	// var cells := util.Cell
	c.events <- CellFlipped{0, util.Cell{X: 0, Y: 0}}
	world := newworld
	// TODO: Execute all turns of the Game of Life.
	for turn = 0; turn < p.Turns; turn++ {
		world = calculateNextState(p, world)
		aliveCells = calculateAliveCells(p, world)
		c.events <- AliveCellsCount{turn, len(aliveCells)}
		c.events <- TurnComplete{turn}
	}
	// c.ioOutput <- world
	// c.events <- FinalTurnComplete{turn,world}

	// TODO: Send correct Events when required, e.g. CellFlipped, TurnComplete and FinalTurnComplete.
	//		 See event.go for a list of all events.

	// Make sure that the Io has finished any output before exiting.

	//TODO na ftiaksoume to output

	// c.events <- ImageOutputComplete{turn,imageName}

	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{turn, Quitting}
	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}
