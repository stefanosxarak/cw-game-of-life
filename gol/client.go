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

// struct to contain quit and other variables
type Client struct {
	quit bool
}

//error handling for server/client
func handleError(err error) {
	if err != nil {
		panic(err)
	}
}

// recieve alive cells from server
func (client *Client) alive(p stubs.Parameters, server *rpc.Client, c clientChannels) {
	args := new(stubs.Default)
	reply := new(stubs.AliveCell)
	server.Call(stubs.AliveCells, args, reply)
	c.events <- AliveCellsCount{reply.Turn, reply.Num}
}

// Gets a proccessed world from server
func (client *Client) worldFromServer(server *rpc.Client) [][]uint8 {
	args := new(stubs.Default)
	reply := new(stubs.Request)
	err := server.Call(stubs.WorldFromServer, args, reply)
	handleError(err)
	return reply.World
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

// Save game and continue playing
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

// calculate alivecells for client
func calculateAlive(world [][]uint8) []util.Cell {
	aliveCells := make([]util.Cell, 0)
	for row := range world {
		for col := range world[row] {
			if world[row][col] == 255 {
				aliveCells = append(aliveCells, util.Cell{X: col, Y: row})
			}
		}
	}
	return aliveCells
}

func (client *Client) keyControl(c clientChannels, p stubs.Parameters, turn int, quit bool, server *rpc.Client) bool {
	//s to save, q to client.quit, p to pause/unpause, k to stop all comms with server
	select {
	case x := <-c.keyPresses:
		if x == 's' {
			fmt.Println("Saving...")
			world := client.worldFromServer(server)
			saveWorld(c, p, turn, world)
		} else if x == 'q' {
			client.quit = true
			c.events <- StateChange{turn, Quitting}
		} else if x == 'p' {
			pause(c, turn, x)

		} else if x == 'k' {
			fmt.Println("Contact with server ceases to exist...")
			world := client.worldFromServer(server)
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

//Client runs the game from the server
func (client *Client) gameExecution(c clientChannels, p stubs.Parameters, server *rpc.Client) (turn int) {

	world := client.worldFromServer(server)

	//Game of Life.
	client.quit = false
	ticker := time.NewTicker(2 * time.Second)
	for turn := 0; turn < p.Turns && client.quit == false; turn++ {
		client.quit = client.keyControl(c, p, turn, client.quit, server)
		world = client.worldFromServer(server)

		//ticker
		select {
		case <-ticker.C:
			client.alive(p, server, c)
		default:
			break
		}

		//update events
		c.events <- TurnComplete{turn}
	}
	//terminate ticker
	ticker.Stop()

	c.events <- FinalTurnComplete{turn, calculateAlive(world)}

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
