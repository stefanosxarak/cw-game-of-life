package gol

import (
	"time"

	"github.com/ChrisGora/semaphore"
	"uk.ac.bris.cs/gameoflife/stubs"
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
type worker struct {
	work  semaphore.Semaphore
	space semaphore.Semaphore
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

// pause the game
// func pause(c distributorChannels, turn int, x rune) {
// 	c.events <- StateChange{turn, Paused}
// 	fmt.Println("The current turn is being processed.")
// 	x = ' '
// 	for x != 'p' {
// 		x = <-c.keyPresses
// 	}
// 	fmt.Println("Continuing")
// 	c.events <- StateChange{turn, Executing}
// }

// Button control
// func keyControl(c distributorChannels, p Params, turn int, quit bool, world [][]uint8) bool {
// 	//s to save, q to quit, p to pause/unpause, k to stop all comms with server
// 	select {
// 	case x := <-c.keyPresses:
// 		if x == 's' {
// 			saveWorld(c, p, turn, world)
// 		} else if x == 'q' {
// 			quit = true
// 			c.events <- StateChange{turn, Quitting}
// 			break
// 		} else if x == 'p' {
// 			pause(c, turn, x)
// 		} else {
// 			New("Wrong Key. Please press 's' to save, 'q' to quit, 'p' to pause/unpause, 'k' to close all server communications ")
// 		}
// 	default:
// 		break
// 	}
// 	return quit
// }

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p stubs.Parameters, c distributorChannels) {

	//Initialization
	world := makeWorld(p.ImageHeight, p.ImageWidth, c)
	newWorld := makeNewWorld(p.ImageHeight, p.ImageWidth)

	// Start making workers and running them
	// workers := createWorkers(p)
	// go runWorkers()

	//Game of Life.
	var turn int
	ticker := time.NewTicker(2 * time.Second)
	for turn = 0; turn < p.Turns; turn++ {

		newWorld = calculateNextState(p, turn, c, world)

		//we add the newly updated world to the grid we had made
		world = newWorld
		newWorld = makeNewWorld(p.ImageHeight, p.ImageWidth)

	}

}
