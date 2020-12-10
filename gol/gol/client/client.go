package gol

import (
	"fmt"
	"net/rpc"

	"uk.ac.bris.cs/gameoflife/util"
)

//TODO:
// Implement a basic controller which can tell the logic engine to evolve Game of Life for the number of turns
// saveWorld
// keyControl
// Implement key k at keyControl
// stubs

type ClientChans struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioInput    <-chan uint8
	ioOutput   chan<- uint8
	keyPresses <-chan rune
}

//error handling for server/client
func handleError(err error) {
	if err != nil {
		panic(err)
	}
}

func calculateAliveClient(p Params, world [][]byte) []util.Cell {
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

//progress to next state and update CellFlipped event
func calculateNextState(p Params, turn int, c ClientChans, world [][]byte) [][]byte {
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

// creates a grid for the current state of the world
func makeWorld(height int, width int, c ClientChans) [][]uint8 {
	world := make([][]uint8, height)
	for row := 0; row < height; row++ {
		world[row] = make([]uint8, width)
		for cell := 0; cell < width; cell++ {
			world[row][cell] = <-c.ioInput
		}
	}
	return world
}

// pause the game
func pause(c ClientChans, turn int, x rune) {
	c.events <- StateChange{turn, Paused}
	fmt.Println("The current turn is being processed.")
	x = ' '
	for x != 'p' {
		x = <-c.keyPresses
	}
	fmt.Println("Continuing")
	c.events <- StateChange{turn, Executing}
}

func keyControl(c ClientChans, p Params, turn int, quit bool, world [][]uint8) bool {
	//s to save, q to quit, p to pause/unpause, k to stop all comms with server
	select {
	case x := <-c.keyPresses:
		if x == 's' {
			saveWorld(c, p, turn, world)
		} else if x == 'q' {
			quit = true
			c.events <- StateChange{turn, Quitting}
		} else if x == 'p' {
			pause(c, turn, x)
		} else if x == 'k' {

		} else {
			New("Wrong Key. Please press 's' to save, 'q' to quit, 'p' to pause/unpause, 'k' to close all server communications ")
		}
	default:
		break
	}
	return quit
}

func saveWorld(c ClientChans, p Params, turn int, world [][]uint8) {
	c.ioCommand <- ioOutput
	outputFilename := fmt.Sprintf("%vx%vx%v", p.ImageWidth, p.ImageHeight, turn)
	c.ioFilename <- outputFilename
	for row := 0; row < p.ImageHeight; row++ {
		for cell := 0; cell < p.ImageWidth; cell++ {
			c.ioOutput <- world[row][cell]
		}
	}
}

func gameExecution(c ClientChans, p Params) {

	//Initialization
	world := makeWorld(p.ImageHeight, p.ImageWidth, c)
	newWorld := makeNewWorld(p.ImageHeight, p.ImageWidth)

	// A variable to store current alive cells
	aliveCells := make([]util.Cell, 0)
	// For all initially alive cells send a CellFlipped Event.
	c.events <- CellFlipped{0, util.Cell{X: 0, Y: 0}}

	//Game of Life.
	quit := false
	var turn int

	for turn = 0; turn < p.Turns && quit == false; turn++ {

		// Waiting all workers to finish each round
		// for _, w := range workers {
		// 	w.space.Wait()
		// }

		quit = keyControl(c, p, turn, quit, world)
		newWorld = calculateNextState(p, turn, c, world)
		aliveCells = calculateAliveCells(p, world)
		//we add the newly updated world to the grid we had made
		world = newWorld
		newWorld = makeNewWorld(p.ImageHeight, p.ImageWidth)

		//update events
		c.events <- AliveCellsCount{turn, len(aliveCells)}
		c.events <- TurnComplete{turn}

		// Workers to start the next round if no q is pressed
		// for j := 0; j < p.Threads && quit == false; j++ {
		// 	workers[j].work.Post()
		// }

	}
	//terminate ticker
	// t.done <- true
}

func clientRun(p Params, c ClientChans) {

	// Establish contact with the server
	srvrAddr := "localhost:8030"
	server, err := rpc.Dial("tcp", srvrAddr)
	handleError(err)
	defer server.Close()
	// err = server.Call(, args, reply)
	handleError(err)

	//fmt.Fprintln(conn, *addrPtr)
	// go read(&conn)

	c.ioCommand <- ioCheckIdle
	<-c.ioIdle
	c.events <- StateChange{turn, Quitting}

	// // Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)

	//TODO Start asynchronously reading and displaying messages
	//TODO Start getting and sending user messages.
}

// func read(conn *net.Conn) {
// 	// In a continuous loop, read a message from the server and display it.
// 	for {
// 		reader := bufio.NewReader(*conn)
// 		msg, _ := reader.ReadString('\n')
// 		fmt.Printf(msg)
// 	}
// }
