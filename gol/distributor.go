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

// type worker struct {
// 	world    [][]uint8
// 	newWorld [][]uint8
// 	startRow int
// 	endRow   int
// }

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
func calculateNextState(start int, end int, p Params, world [][]byte, c distributorChannels, turn int) [][]byte {
	rows := end - start
	newWorld := makeNewWorld(rows, p.ImageWidth)
	i := 0
	for y := start; y < end; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			neighbours := calculateNeighbours(p, x, y, world)
			if world[y][x] == alive {
				if neighbours == 2 || neighbours == 3 {
					newWorld[i][x] = alive
				} else {
					newWorld[i][x] = dead
					c.events <- CellFlipped{turn, util.Cell{X: x, Y: y}}
				}
			} else {
				if neighbours == 3 {
					newWorld[i][x] = alive
					c.events <- CellFlipped{turn, util.Cell{X: x, Y: y}}
				} else {
					newWorld[i][x] = dead
				}
			}
		}
		i++
	}
	return newWorld
}

//calculateAliveCells function takes the world as input and returns the (x, y) coordinates of all the cells that are alive.
func calculateAliveCells(world [][]byte) []util.Cell {
	aliveCells := []util.Cell{}

	for row := range world {
		for col := range world[row] {
			if world[row][col] == alive {
				aliveCells = append(aliveCells, util.Cell{X: col, Y: row})
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
			c.events <- CellFlipped{0, util.Cell{X: cell, Y: row}}
		}
	}
	return world
}

//A 2D slice to store the updated world.
func makeNewWorld(height int, width int) [][]byte {
	newWorld := make([][]byte, height)
	for row := 0; row < height; row++ {
		newWorld[row] = make([]byte, width)
	}
	return newWorld
}

// worker functions
// func (w *worker) createWorkers(p Params) {
// 	rowsPerWorker := p.ImageHeight / p.Threads
// 	remaining := p.ImageHeight % p.Threads
// 	startRow := 0
// 	for i := 0; i < p.Threads; i++ {
// 		workerRows := rowsPerWorker
// 		//adds one of the remaining rows to a worker
// 		if remaining > 0 {
// 			workerRows++
// 			remaining--
// 		}
// 		w := worker{}
// 		w.startRow = startRow
// 		w.endRow = startRow + workerRows - 1
// 	}
	// return startRow,endRow
// }

func runWorkers(startRow int, endRow int, world [][]byte, p Params,  c distributorChannels, turn int, slice chan<- [][]byte) {
	// Implement the worker and calculateNextState
	worldPart := calculateNextState(startRow, endRow, p, world, c, turn)
	slice <- worldPart
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

	//Input data
	c.ioCommand <- ioInput
	imageName := fmt.Sprintf("%dx%d", p.ImageHeight, p.ImageWidth)
	c.ioFilename <- imageName

	//Initialization
	newWorld := makeNewWorld(p.ImageHeight, p.ImageWidth)

	// The io goroutine sends the requested image byte by byte, in rows.
	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			val := <-c.ioInput
			if val != 0 {
				newWorld[y][x] = val
			}
		}
	}

	// Start making workers and running them
	rowsPerWorker := p.ImageHeight / p.Threads
	remaining := p.ImageHeight % p.Threads

	// A variable to store current alive cells
	aliveCells := calculateAliveCells(newWorld)

	//create a channel to store the slices of the world for each worker
	slice := make([]chan [][]byte, p.Threads)
	for i := range slice {
		slice[i] = make(chan [][]byte)
	}

	// For all initially alive cells send a CellFlipped Event.
	c.events <- CellFlipped{0, util.Cell{X: 0, Y: 0}}

	//Game of Life.
	quit := false
	var turn int
	ticker := time.NewTicker(2 * time.Second)
	for turn = 0; turn < p.Turns && quit == false; turn++ {

		// newWorld = calculateNextState(p, turn, c, world)
		aliveCells = calculateAliveCells(newWorld)
		if p.Threads > 1 {
			for i := 0; i < p.Threads; i++ {
				//if the number of threads doesnt divide well with the image, then
				//if the current thread is the last one, give it the remaining lines to calculate
				if (remaining > 0) && ((i + 1) == p.Threads) {
					go runWorkers(i*rowsPerWorker, ((i+1)*rowsPerWorker)+remaining, newWorld, p, c, turn, slice[i])
				} else { //else, just give each thread the appointed rowsPerWorker
					go runWorkers(i*rowsPerWorker, (i+1)*rowsPerWorker, newWorld, p, c, turn, slice[i])
				}
			}
			//we add the newly updated world to the grid we had made

			// A temprorary world to append all the slices to.
			temp := make([][]byte, 0)
			for i := 0; i < p.Threads; i++ {
				part := <-slice[i]
				temp = append(temp, part...)
			}

			for y := 0; y < p.ImageHeight; y++ {
				for x := 0; x < p.ImageWidth; x++ {
					// Swap temporary world with the real newWorld[y][x]
					newWorld[y][x] = temp[y][x]
				}
			}

		} else {
			//if the number of worker threads is one,
			//give the lone worker the whole image to calculate

			start := 0
			end := p.ImageHeight

			newWorld = calculateNextState(start, end, p, newWorld, c, turn)
			c.events <- TurnComplete{CompletedTurns: turn}
		}

		// world = newWorld
		// newWorld = makeNewWorld(p.ImageHeight, p.ImageWidth)

		quit = keyControl(c, p, turn, quit, newWorld)
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
	saveWorld(c, p, turn, newWorld)

	// Make sure that the Io has finished any output before exiting.
	c.events <- FinalTurnComplete{turn, calculateAliveCells(newWorld)}
	c.events <- ImageOutputComplete{turn, imageName}

	c.ioCommand <- ioCheckIdle
	<-c.ioIdle
	c.events <- StateChange{turn, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}
