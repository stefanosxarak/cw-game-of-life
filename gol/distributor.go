package gol

import (
	"fmt"
	"time"

	"uk.ac.bris.cs/gameoflife/util"
)

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioOutput   chan<- uint8
	ioInput    <-chan uint8
	keyPresses <-chan rune
}

//error struct
type errorString struct {
	s string
}

const alive = 255
const dead = 0

func mod(x, m int) int {
	return (x + m) % m
}

//error handling
func (e *errorString) Error() string {
	return e.s
}
func New(text string) error {
	return &errorString{text}
}

//calculateNeighbors takes the current state of the world and completes one evolution of the world. It then returns the result
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

//progress to next state
func calculateNextState(p Params, turn int, c distributorChannels, world [][]byte) [][]byte {
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
					c.events <- CellFlipped{turn, util.Cell{X: x, Y: y}}
				}
			} else {
				if neighbours == 3 {
					newWorld[y][x] = alive
					c.events <- CellFlipped{turn, util.Cell{X: x, Y: y}}
				} else {
					newWorld[y][x] = dead
				}
			}
		}
	}
	return newWorld
}

//calculateAliveCells function takes the world as input and returns the (x, y) coordinates of all the cells that are alive.
func calculateAliveCells(p Params, world [][]byte) []util.Cell {
	aliveCells := []util.Cell{}

	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			if world[y][x] == alive {
				aliveCells = append(aliveCells, util.Cell{X: x, Y: y})
			}
		}
	}

	return aliveCells
}

// creates a grid for the current state of the world
func makeWorld(height int, width int, c distributorChannels) [][]uint8 {
	world := make([][]uint8, height)
	for row := 0; row < height; row++ {
		world[row] = make([]uint8, width)
		for cell := 0; cell < width; cell++ {
			world[row][cell] = <-c.ioInput
		}
	}
	return world
}

//A 2D slice to store the updated world.
func makeNewWorld(height int, width int) [][]uint8 {
	newWorld := make([][]uint8, height)
	for row := 0; row < height; row++ {
		newWorld[row] = make([]uint8, width)
	}
	return newWorld
}

// workers
func createWorkers(numOfWorkers int, rowsPerWorker int) {
	for i := 0; i < numOfWorkers; i++ {
		// workerRows := rowsPerWorker
	}
}

func runWorkers() {}

func saveWorld(c distributorChannels, p Params, turn int, world [][]uint8) {
	c.ioCommand <- ioOutput
	outputFilename := fmt.Sprintf("%vx%vx%v", p.ImageWidth, p.ImageHeight, turn)
	c.ioFilename <- outputFilename
	for row := 0; row < p.ImageHeight; row++ {
		for cell := 0; cell < p.ImageWidth; cell++ {
			c.ioOutput <- world[row][cell]
		}
	}
}

// pause the game
func pause(c distributorChannels, turn int, x rune) {
	c.events <- StateChange{turn, Paused}
	fmt.Println("The current turn is being processed.")
	x = ' '
	for x != 'p' {
		x = <-c.keyPresses
	}
	fmt.Println("Continuing")
	c.events <- StateChange{turn, Executing}
}

// Button control
func keyControl(c distributorChannels, p Params, turn int, quit bool, world [][]uint8) bool {
	//s to save, q to quit, p to pause/unpause
	select {
	case x := <-c.keyPresses:
		if x == 's' {
			saveWorld(c, p, turn, world)
		} else if x == 'q' {
			quit = true
			c.events <- StateChange{turn, Quitting}
		} else if x == 'p' {
			pause(c, turn, x)
		}
	default:
		break
	}
	return quit
}

// ticker function
func ticker(c distributorChannels, turn int, aliveCells []util.Cell, done <-chan bool) {
	ticker := time.NewTicker(2 * time.Second)
	done = make(chan bool)

	for {
		select {
		case t := <-ticker.C:
			c.events <- AliveCellsCount{turn, len(aliveCells)}
			fmt.Println("Tick at", t)
		case <-done:
			return
		}
	}
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {
	//Input data
	c.ioCommand <- ioInput
	imageName := fmt.Sprintf("%dx%d", p.ImageHeight, p.ImageWidth)
	c.ioFilename <- imageName

	//Initialization
	world := makeWorld(p.ImageHeight, p.ImageWidth, c)
	newWorld := makeNewWorld(p.ImageHeight, p.ImageWidth)

	// rowsPerWorker := p.ImageHeight / p.Threads


	//ticker variables
	// ticker := time.NewTicker(2 * time.Second)
	// done := make(chan bool)

	// A variable to store current alive cells
	aliveCells := calculateAliveCells(p, newWorld)

	// For all initially alive cells send a CellFlipped Event.
	c.events <- CellFlipped{0, util.Cell{X: 0, Y: 0}}

	//Game of Life.
	quit := false
	turn := 0
	for turn = 0; turn < p.Turns && quit == false; turn++ {

		// wait for all workers to complete this turn
		// for _, w := range workers {
		// 	w.space.Wait()
		// }

		// for i := 0; i < p.Threads; i++ {
		// 	x := <-wc.syncChan[i]
		// 	if x != turn {
		// 		New("Out of sync")
		// 	}
		// }

		keyControl(c, p, turn, quit, world)
		newWorld = calculateNextState(p, turn, c, world)
		aliveCells = calculateAliveCells(p, world)
		// go ticker(c,turn, aliveCells, done)
		

		//update events
		c.events <- AliveCellsCount{turn, len(aliveCells)}
		c.events <- TurnComplete{turn}

		//we add the newly updated world to the grid we had made
		world = newWorld
		newWorld = makeNewWorld(p.ImageHeight, p.ImageWidth)
	}
	// ticker.Stop()
	// done <- true

	//saves the result to a file
	saveWorld(c, p, turn, world)
	
	// Make sure that the Io has finished any output before exiting.
	c.events <- FinalTurnComplete{turn, calculateAliveCells(p, world)}
	c.events <- ImageOutputComplete{turn, imageName}
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle
	c.events <- StateChange{turn, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}
