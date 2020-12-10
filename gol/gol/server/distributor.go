package gol

import (
	"github.com/ChrisGora/semaphore"
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

//A 2D slice to store the updated world.
func makeNewWorld(height int, width int) [][]uint8 {
	newWorld := make([][]uint8, height)
	for row := 0; row < height; row++ {
		newWorld[row] = make([]uint8, width)
	}
	return newWorld
}

// worker functions
func (w *worker) createWorkers(p Params) []worker {
	rowsPerWorker := p.ImageHeight / p.Threads
	remaining := p.ImageHeight % p.Threads
	workers := make([]worker, p.Threads)
	for i := 0; i < p.Threads; i++ {
		workers[i] = worker{}
		workerRows := rowsPerWorker

		//adds one of the remaining rows to a worker
		if remaining > 0 {
			workerRows++
			remaining--
		}

		//Semaphores with buffer size 1 so all workers keep up
		w.work = semaphore.Init(1, 1)
		w.space = semaphore.Init(1, 0)

	}
	return workers
}

func (w *worker) runWorkers(p Params, c distributorChannels, world [][]uint8) {
	//TODO :
	// Implement the worker and calculateNextState
	for turn := 0; ; turn++ {
		w.work.Wait()
		// workerSlice := calculateNextState(p, turn, c, world)
		w.space.Post()
	}
}

// distributor divides the work between workers and interacts with other goroutines.
func (w *worker) distributor(p Params, c distributorChannels) {

	// Start making workers and running them
	workers := createWorkers(p)
	go runWorkers()

	//ticker channels
	t := TickerChans{}
	t.tick = make(chan bool)
	t.done = make(chan bool)

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
