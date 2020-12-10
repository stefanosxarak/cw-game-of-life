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

//Global Variables
const alive = 255
const dead = 0

//error handling for server/client
func handleError(err error) {
	if err != nil {
		panic(err)
	}
}

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

//A 2D slice to store the updated world.
func makeNewWorld(height int, width int) [][]uint8 {
	newWorld := make([][]uint8, height)
	for row := 0; row < height; row++ {
		newWorld[row] = make([]uint8, width)
	}
	return newWorld
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

func gameExecution(c ClientChans, p Params) (turn int) {

	//Initialization
	world := makeWorld(p.ImageHeight, p.ImageWidth, c)
	newWorld := makeNewWorld(p.ImageHeight, p.ImageWidth)

	// A variable to store current alive cells
	aliveCells := make([]util.Cell, 0)
	// For all initially alive cells send a CellFlipped Event.
	c.events <- CellFlipped{0, util.Cell{X: 0, Y: 0}}

	//Game of Life.
	quit := false
	for turn = 0; turn < p.Turns && quit == false; turn++ {

		quit = keyControl(c, p, turn, quit, world)
		newWorld = calculateNextState(p, turn, c, world)
		aliveCells = calculateAliveCells(p, world)
		//we add the newly updated world to the grid we had made
		world = newWorld
		newWorld = makeNewWorld(p.ImageHeight, p.ImageWidth)

		//update events
		c.events <- AliveCellsCount{turn, len(aliveCells)}
		c.events <- TurnComplete{turn}

	}
	//terminate ticker
	// t.done <- true
	c.events <- FinalTurnComplete{turn, calculateAliveCells(p, world)}
	

	//saves the result to a file
	saveWorld(c, p, turn, world)

	return turn
}

func clientRun(p Params, c ClientChans, server *rpc.Client) {

	// Establish contact with the server
	srvrAddr := "localhost:8030"
	server, err := rpc.Dial("tcp", srvrAddr)
	handleError(err)
	defer server.Close()
	// err = server.Call(, args, reply)
	handleError(err)

	turn := gameExecution(c, p)
	server.Close()

	c.ioCommand <- ioCheckIdle
	<-c.ioIdle
	c.events <- StateChange{turn, Quitting}

	// // Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)

}

// func read(conn *net.Conn) {
// 	// In a continuous loop, read a message from the server and display it.
// 	for {
// 		reader := bufio.NewReader(*conn)
// 		msg, _ := reader.ReadString('\n')
// 		fmt.Printf(msg)
// 	}
// }
