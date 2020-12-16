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

type worker struct {
	world    [][]uint8
	newWorld [][]uint8
	startRow int
	endRow   int
}

const alive = 255
const dead = 0

func mod(x, m int) int {
	return (x + m) % m
}

//error handling for distributor
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

//progress to next state and update CellFlipped event
func (w *worker) calculateNextState(start int, end int, p Params, world [][]byte, c distributorChannels, turn int) {
	rows := end - start
	w.newWorld = make([][]uint8, rows)
	for i := 0; i < rows; i++ {
		w.newWorld[i] = make([]uint8, p.ImageWidth)
	}

	i := 0
	for y := start; y < end; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			neighbours := calculateNeighbours(p, x, y, world)
			if w.world[y][x] == alive {
				if neighbours == 2 || neighbours == 3 {
					w.newWorld[i][x] = alive
				} else {
					w.newWorld[i][x] = dead
					c.events <- CellFlipped{turn, util.Cell{X: x, Y: y}}
				}
			} else {
				if neighbours == 3 {
					w.newWorld[i][x] = alive
					c.events <- CellFlipped{turn, util.Cell{X: x, Y: y}}
				} else {
					w.newWorld[i][x] = dead
				}
			}
		}
		i++
	}

}

//calculateAliveCells function takes the world as input and returns the (x, y) coordinates of all the cells that are alive.
func (w *worker) calculateAliveCells(p Params) []util.Cell {
	aliveCells := []util.Cell{}

	for row := range w.world {
		for col := range w.world[row] {
			if w.world[row][col] == alive {
				aliveCells = append(aliveCells, util.Cell{X: col, Y: row})
			}
		}
	}

	return aliveCells
}

// creates a grid for the current state of the world
func (w *worker) makeWorld(height int, width int, c distributorChannels) {
	w.world = make([][]uint8, height)
	for row := 0; row < height; row++ {
		w.world[row] = make([]uint8, width)
		for cell := 0; cell < width; cell++ {
			w.world[row][cell] = <-c.ioInput
		}
	}

}

//A 2D slice to store the updated world.
func (w *worker) makeNewWorld(height int, width int) {
	w.newWorld = make([][]uint8, height)
	for row := 0; row < height; row++ {
		w.newWorld[row] = make([]uint8, width)
	}
}

// worker functions
func (w *worker) createWorkers(p Params) {
	rowsPerWorker := p.ImageHeight / p.Threads
	remaining := p.ImageHeight % p.Threads
	startRow := 0
	for i := 0; i < p.Threads; i++ {
		workerRows := rowsPerWorker
		//adds one of the remaining rows to a worker
		if remaining > 0 {
			workerRows++
			remaining--
		}
		w := worker{}
		w.startRow = startRow
		w.endRow = startRow + workerRows - 1
	}
	// return startRow,endRow
}

func (w *worker) runWorkers(startRow int, endRow int, p Params, turn int, c distributorChannels, world [][]byte) {
	// Implement the worker and calculateNextState
	w.calculateNextState(startRow, endRow, p, world, c, turn)

}

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
	//s to save, q to quit, p to pause/unpause, k to stop all comms with server
	select {
	case x := <-c.keyPresses:
		if x == 's' {
			saveWorld(c, p, turn, world)
		} else if x == 'q' {
			quit = true
			c.events <- StateChange{turn, Quitting}
			break
		} else if x == 'p' {
			pause(c, turn, x)
		} else {
			New("Wrong Key. Please press 's' to save, 'q' to quit, 'p' to pause/unpause, 'k' to close all server communications ")
		}
	default:
		break
	}
	return quit
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {
	w := worker{}

	//Input data
	c.ioCommand <- ioInput
	imageName := fmt.Sprintf("%dx%d", p.ImageHeight, p.ImageWidth)
	c.ioFilename <- imageName

	//Initialization
	w.makeWorld(p.ImageHeight, p.ImageWidth, c)
	w.makeNewWorld(p.ImageHeight, p.ImageWidth)

	// Start making workers and running them
	w.createWorkers(p)

	// A variable to store current alive cells
	aliveCells := w.calculateAliveCells(p)

	// For all initially alive cells send a CellFlipped Event.
	c.events <- CellFlipped{0, util.Cell{X: 0, Y: 0}}

	//Game of Life.
	quit := false
	var turn int
	ticker := time.NewTicker(2 * time.Second)
	for turn = 0; turn < p.Turns && quit == false; turn++ {

		// newWorld = calculateNextState(p, turn, c, world)
		aliveCells = w.calculateAliveCells(p)

		go w.runWorkers(w.startRow, w.endRow, p, turn, c, w.world)
		
		//we add the newly updated world to the grid we had made
		w.world = w.newWorld
		w.makeNewWorld(p.ImageHeight, p.ImageWidth)

		quit = keyControl(c, p, turn, quit, w.world)
		//ticker
		select {
		case <-ticker.C:
			c.events <- AliveCellsCount{turn, len(aliveCells)}
		default:
			break
		}

		//update turns
		c.events <- TurnComplete{turn}

	}
	//terminate ticker
	ticker.Stop()

	//saves the result to a file
	saveWorld(c, p, turn, w.world)

	// Make sure that the Io has finished any output before exiting.
	c.events <- FinalTurnComplete{turn, w.calculateAliveCells(p)}
	c.events <- ImageOutputComplete{turn, imageName}

	c.ioCommand <- ioCheckIdle
	<-c.ioIdle
	c.events <- StateChange{turn, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}
