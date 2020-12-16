package gol

import (
	"fmt"
	"log"
	"net/rpc"
	"time"

	"uk.ac.bris.cs/gameoflife/stubs"

	"uk.ac.bris.cs/gameoflife/util"
)

type clientChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioInput    <-chan uint8
	ioOutput   chan<- uint8
	keyPresses <-chan rune
}

//TODO:
// Implement a basic controller which can tell the logic engine to evolve Game of Life for the number of turns
// keyControl
// Implement key k at keyControl
// stubs

type Client struct {
	quit bool
}

const alive = 255
const dead = 0

//error handling for server/client
func handleError(err error) {
	if err != nil {
		panic(err)
	}
}

func calculateAliveCells(p stubs.Parameters, world [][]byte) []util.Cell {
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

// Gets a proccessed world from server
func (client *Client) worldFromServer(server *rpc.Client) (world [][]uint8, turn int) {
	args := new(stubs.Default)
	reply := new(stubs.Request)
	err := server.Call(stubs.WorldFromServer, args, reply)
	if err != nil {
		fmt.Println("err", err)
	}
	return reply.World, reply.Param.Turns
}

// Terminate contact with server
func (client *Client) killServer(server *rpc.Client) (turn int) {
	args := new(stubs.Default)
	reply := new(stubs.Request)
	err := server.Call(stubs.Kill, args, reply)
	handleError(err)
	return reply.Param.Turns
}

// pause game proccessing 
func pause(c clientChannels, turn int, x rune) {
	c.events <- StateChange{turn, Paused}
	fmt.Println("The current turn is being processed.")
	x = ' '
	for x != 'p' {
		x = <-c.keyPresses
	}
	fmt.Println("Continuing")
	c.events <- StateChange{turn, Executing}
}

func saveWorld(c clientChannels, p stubs.Parameters, turn int, world [][]uint8) {
	c.ioCommand <- ioOutput
	outputFilename := fmt.Sprintf("%vx%vx%v", p.ImageWidth, p.ImageHeight, turn)
	c.ioFilename <- outputFilename
	for row := 0; row < p.ImageHeight; row++ {
		for cell := 0; cell < p.ImageWidth; cell++ {
			c.ioOutput <- world[row][cell]
		}
	}
	// Notify events every time the world is saved i.e. every time this function is called
	c.events <- ImageOutputComplete{turn, outputFilename}
}

func (client *Client) keyControl(c clientChannels, p stubs.Parameters, turn int, quit bool, server *rpc.Client) bool {
	//s to save, q to client.quit, p to pause/unpause, k to stop all comms with server
	select {
	case x := <-c.keyPresses:
		if x == 's' {
			fmt.Println("Saving...")
			world, current := client.worldFromServer(server)
			turn = current
			saveWorld(c, p, turn, world)
		} else if x == 'q' {
			client.quit = true
			c.events <- StateChange{turn, Quitting}
		} else if x == 'p' {
			pause(c, turn, x)

		} else if x == 'k' {
			fmt.Println("Contact with server ceases to exist...")
			world, current := client.worldFromServer(server)
			turn = current
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

//TODO needs to be safely transfered to server
func (client *Client) gameExecution(c clientChannels, p stubs.Parameters, server *rpc.Client) (turn int) {

	//Initialization
	// world := makeWorld(p.ImageHeight, p.ImageWidth, c)
	// newWorld := makeNewWorld(p.ImageHeight, p.ImageWidth)

	// A variable to store current alive cells
	aliveCells := make([]util.Cell, 0)

	// For all initially alive cells send a CellFlipped Event.
	c.events <- CellFlipped{0, util.Cell{X: 0, Y: 0}}

	world, current := client.worldFromServer(server)
	turn = current

	//Game of Life.
	client.quit = false
	ticker := time.NewTicker(2 * time.Second)
	for turn = 0; turn < p.Turns && client.quit == false; turn++ {

		client.quit = client.keyControl(c, p, turn, client.quit, server)

		//ticker
		select {
		case <-ticker.C:
			aliveCells = calculateAliveCells(p, world)
			c.events <- AliveCellsCount{turn, len(aliveCells)}

		default:
			break
		}

		//update events
		c.events <- TurnComplete{turn}

	}
	//terminate ticker
	ticker.Stop()
	c.events <- FinalTurnComplete{turn, calculateAliveCells(p, world)}

	//saves the result to a file
	saveWorld(c, p, turn, world)

	return turn
}

func (client *Client) clientRun(p stubs.Parameters, c clientChannels, server *rpc.Client) {

	//Extract final info and close conn with server
	turn := client.gameExecution(c, p, server)
	if client.quit == true {
		client.killServer(server)
	}
	server.Close()

	// Exit
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle
	c.events <- StateChange{turn, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)

}
