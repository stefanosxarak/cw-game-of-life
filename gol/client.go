package gol

import (
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

// func keyControl(c distributorChannels, p Params, turn int, quit bool, world [][]uint8) bool {
// 	//s to save, q to quit, p to pause/unpause, k to stop all comms with server
// 	select {
// 	case x := <-c.keyPresses:
// 		if x == 's' {
// 			saveWorld(c, p, turn, world)
// 		} else if x == 'q' {
// 			quit = true
// 			c.events <- StateChange{turn, Quitting}
// 		} else if x == 'p' {
// 			pause(c, turn, x)
// 		} else if x == 'k' {

// 		} else {
// 			New("Wrong Key. Please press 's' to save, 'q' to quit, 'p' to pause/unpause, 'k' to close all server communications ")
// 		}
// 	default:
// 		break
// 	}
// 	return quit
// }

// func saveWorld(c distributorChannels, p Params, turn int, world [][]uint8) {
// 	c.ioCommand <- ioOutput
// 	outputFilename := fmt.Sprintf("%vx%vx%v", p.ImageWidth, p.ImageHeight, turn)
// 	c.ioFilename <- outputFilename
// 	for row := 0; row < p.ImageHeight; row++ {
// 		for cell := 0; cell < p.ImageWidth; cell++ {
// 			c.ioOutput <- world[row][cell]
// 		}
// 	}
// }

// func read(conn *net.Conn) {
// 	// In a continuous loop, read a message from the server and display it.
// 	for {
// 		reader := bufio.NewReader(*conn)
// 		msg, _ := reader.ReadString('\n')
// 		fmt.Printf(msg)
// 	}
// }

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

	// c.ioCommand <- ioCheckIdle
	// <-c.ioIdle
	// c.events <- StateChange{turn, Quitting}

	// // Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	// close(c.events)

	//TODO Start asynchronously reading and displaying messages
	//TODO Start getting and sending user messages.
}
