package gol

import (
	"bufio"
	"flag"
	"fmt"
	"net"
)

//TODO:
// Implement a basic controller which can tell the logic engine to evolve Game of Life for the number of turns
// saveWorld
// keyControl
// Implement key k at keyControl

type ClientChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioInput    <-chan uint8
	ioOutput   chan<- uint8
	keyPresses <-chan rune
}

func read(conn *net.Conn) {
	// In a continuous loop, read a message from the server and display it.
	for {
		reader := bufio.NewReader(*conn)
		msg, _ := reader.ReadString('\n')
		fmt.Printf(msg)
	}
}

func client(p Params, c ClientChannels) {

	// Get the server address and port from the commandline arguments.
	addrPtr := flag.String("ip", "100.84.31.131:8030", "IP:port string to connect to")
	flag.Parse()
	//TODO Try to connect to the server
	conn, err := net.Dial("tcp", *addrPtr)
	handleError(err)

	// rpcConn, err := rpc.Dial("tcp", "127.0.0.1:8030")
	handleError(err)

	//fmt.Fprintln(conn, *addrPtr)
	go read(&conn)

	//TODO Start asynchronously reading and displaying messages
	//TODO Start getting and sending user messages.
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
