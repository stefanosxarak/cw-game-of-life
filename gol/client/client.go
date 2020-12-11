package gol

import (
	"fmt"
	"log"
	"net/rpc"

	"uk.ac.bris.cs/gameoflife/stubs"

	"uk.ac.bris.cs/gameoflife/util"
)

//TODO:
// Implement a basic controller which can tell the logic engine to evolve Game of Life for the number of turns
// saveWorld
// keyControl
// Implement key k at keyControl
// stubs

type clientChans struct {
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

type Client struct {
	// t    Ticker
	quit bool
}

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
func calculateNextState(p Params, turn int, c clientChans, world [][]byte) [][]byte {
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

// Gets a proccessed world from server
func worldFromServer(server *rpc.Client) (world [][]uint8) {
	args := new(stubs.Default)
	reply := new(stubs.Request)
	err := server.Call(stubs.worldFromServer, args, reply)
	handleError(err)
	return reply.World
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
func makeWorld(height int, width int, c clientChans) [][]uint8 {
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

// pause the game and not the server!
func pause(c clientChans, turn int, x rune) {
	c.events <- StateChange{turn, Paused}
	fmt.Println("The current turn is being processed.")
	x = ' '
	for x != 'p' {
		x = <-c.keyPresses
	}
	fmt.Println("Continuing")
	c.events <- StateChange{turn, Executing}
}

func (client *Client) keyControl(c clientChans, p Params, turn int, quit bool, server *rpc.Client) bool {
	//s to save, q to client.quit, p to pause/unpause, k to stop all comms with server
	select {
	case x := <-c.keyPresses:
		if x == 's' {
			world := worldFromServer(server)
			saveWorld(c, p, turn, world)
		} else if x == 'q' {
			client.quit = true
			c.events <- StateChange{turn, Quitting}
		} else if x == 'p' {
			pause(c, turn, x)

		} else if x == 'k' {
			world := worldFromServer(server)
			client.quit = true
			saveWorld(c, p, turn, world)
			c.events <- StateChange{turn, Quitting}

		} else {
			log.Fatalf("Wrong Key. Please press 's' to save, 'q' to client.quit, 'p' to pause/unpause, 'k' to close all server communications ")
		}
	default:
		break
	}
	return client.quit
}

func saveWorld(c clientChans, p Params, turn int, world [][]uint8) {
	c.ioCommand <- ioOutput
	outputFilename := fmt.Sprintf("%vx%vx%v", p.ImageWidth, p.ImageHeight, turn)
	c.ioFilename <- outputFilename
	for row := 0; row < p.ImageHeight; row++ {
		for cell := 0; cell < p.ImageWidth; cell++ {
			c.ioOutput <- world[row][cell]
		}
	}
}

func (client *Client) gameExecution(c clientChans, p Params, server *rpc.Client) (turn int) {

	//Initialization
	world := makeWorld(p.ImageHeight, p.ImageWidth, c)
	newWorld := makeNewWorld(p.ImageHeight, p.ImageWidth)

	// A variable to store current alive cells
	aliveCells := make([]util.Cell, 0)

	// For all initially alive cells send a CellFlipped Event.
	c.events <- CellFlipped{0, util.Cell{X: 0, Y: 0}}

	//Game of Life.
	client.quit = false
	for turn = 0; turn < p.Turns && client.quit == false; turn++ {

		client.quit = client.keyControl(c, p, turn, client.quit, server)
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

func (client *Client) clientRun(p Params, c clientChans, server *rpc.Client) {
	// Input data
	c.ioCommand <- ioInput
	imageName := fmt.Sprintf("%dx%d", p.ImageHeight, p.ImageWidth)
	c.ioFilename <- imageName

	// Place ticker here

	//Extract final info and close conn with server

	turn := client.gameExecution(c, p, server)
	if client.quit == true {
		// args := new(stubs.Default)
		// reply := new(stubs.Turn)
		// err := server.Call(stubs.Kill, args, reply)
		// handleError(err)
	}
	server.Close()

	// Make sure that the Io has finished any output before exiting.
	c.events <- ImageOutputComplete{turn, imageName}

	// Exit
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle
	c.events <- StateChange{turn, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)

}
