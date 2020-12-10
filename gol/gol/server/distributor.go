package gol

import (
	"github.com/ChrisGora/semaphore"
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

func mod(x, m int) int {
	return (x + m) % m
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

// func (w *worker) runWorkers(p Params, c distributorChannels, world [][]uint8) {
// 	//TODO :
// 	// Implement the worker and calculateNextState
// 	for turn := 0; ; turn++ {
// 		w.work.Wait()
// 		// workerSlice := calculateNextState(p, turn, c, world)
// 		w.space.Post()
// 	}
// }

// distributor divides the work between workers and interacts with other goroutines.
// func (w *worker) distributor(p Params, c distributorChannels) {

	// Waiting all workers to finish each round
	// for _, w := range workers {
	// 	w.space.Wait()
	// }

	// Start making workers and running them
	// workers := createWorkers(p)
	// go runWorkers()

	//ticker channels
	// t := TickerChans{}
	// t.tick = make(chan bool)
	// t.done = make(chan bool)

	// Workers to start the next round if no q is pressed
	// for j := 0; j < p.Threads && quit == false; j++ {
	// 	workers[j].work.Post()
	// }

	//saves the result to a file
	// saveWorld(c, p, turn, world)

	// Make sure that the Io has finished any output before exiting.
	// c.events <- ImageOutputComplete{turn, imageName}

// }
