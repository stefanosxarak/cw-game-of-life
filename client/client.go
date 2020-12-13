package gol

import (
	"fmt"
	"log"
	"net/rpc"
	"time"

	"uk.ac.bris.cs/gameoflife/util"
)

//TODO:
// Implement a basic controller which can tell the logic engine to evolve Game of Life for the number of turns
// saveWorld
// keyControl
// Implement key k at keyControl
// stubs

type Client struct {
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

// Gets a proccessed world from server
// func worldFromServer(server *rpc.Client) (world [][]uint8) {
// 	args := new(stubs.Default)
// 	reply := new(stubs.Request)
// 	err := server.Call(stubs.worldFromServer, args, reply)
// 	handleError(err)
// 	return reply.World
// }

// pause game proccessing and not the server!
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

func saveWorld(c distributorChannels, p Params, turn int, world [][]uint8) {
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

func (client *Client) keyControl(c distributorChannels, p Params, turn int, quit bool, server *rpc.Client) bool {
	//s to save, q to client.quit, p to pause/unpause, k to stop all comms with server
	select {
	case x := <-c.keyPresses:
		if x == 's' {
			// world := worldFromServer(server)
			saveWorld(c, p, turn, world)
		} else if x == 'q' {
			client.quit = true
			c.events <- StateChange{turn, Quitting}
		} else if x == 'p' {
			pause(c, turn, x)

		} else if x == 'k' {
			// world := worldFromServer(server)
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
func (client *Client) gameExecution(c distributorChannels, p Params, server *rpc.Client) (turn int) {

	//Initialization
	world := makeWorld(p.ImageHeight, p.ImageWidth, c)
	newWorld := makeNewWorld(p.ImageHeight, p.ImageWidth)

	// A variable to store current alive cells
	aliveCells := make([]util.Cell, 0)

	// For all initially alive cells send a CellFlipped Event.
	c.events <- CellFlipped{0, util.Cell{X: 0, Y: 0}}

	//Game of Life.
	client.quit = false
	ticker := time.NewTicker(2 * time.Second)
	for turn = 0; turn < p.Turns && client.quit == false; turn++ {

		newWorld = calculateNextState(p, turn, c, world)
		aliveCells = calculateAliveCells(p, world)

		//we add the newly updated world to the grid we had made
		world = newWorld
		newWorld = makeNewWorld(p.ImageHeight, p.ImageWidth)

		client.quit = client.keyControl(c, p, turn, client.quit, server)

		//ticker
		select {
		case <-ticker.C:
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

func (client *Client) clientRun(p Params, c distributorChannels, server *rpc.Client) {
	// Input data
	c.ioCommand <- ioInput
	imageName := fmt.Sprintf("%dx%d", p.ImageHeight, p.ImageWidth)
	c.ioFilename <- imageName

	//Extract final info and close conn with server
	turn := client.gameExecution(c, p, server)
	if client.quit == true {
		// KILL SERVER

		// args := new(stubs.Default)
		// reply := new(stubs.Turn)
		// err := server.Call(stubs.Kill, args, reply)
		// handleError(err)
	}
	server.Close()

	// Exit
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle
	c.events <- StateChange{turn, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)

}
