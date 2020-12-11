package gol

// "github.com/ChrisGora/semaphore"

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioOutput   chan<- uint8
	ioInput    <-chan uint8
	keyPresses <-chan rune
}

func mod(x, m int) int {
	return (x + m) % m
}

// worker functions
// func (w *worker) createWorkers(p Params) []worker {
// 	rowsPerWorker := p.ImageHeight / p.Threads
// 	remaining := p.ImageHeight % p.Threads
// 	workers := make([]worker, p.Threads)
// 	for i := 0; i < p.Threads; i++ {
// 		workers[i] = worker{}
// 		workerRows := rowsPerWorker

// 		//adds one of the remaining rows to a worker
// 		if remaining > 0 {
// 			workerRows++
// 			remaining--
// 		}

// 		//Semaphores with buffer size 1 so all workers keep up
// 		w.work = semaphore.Init(1, 1)
// 		w.space = semaphore.Init(1, 0)

// 	}
// 	return workers
// }

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
func distributor(p Params, c distributorChannels) {


}
