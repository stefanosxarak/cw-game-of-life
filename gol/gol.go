package gol

import (
	"fmt"
	"net/rpc"

	"uk.ac.bris.cs/gameoflife/stubs"
)

// Params provides the details of how to run the Game of Life and which image to load.
type Params struct {
	Turns       int
	Threads     int
	ImageWidth  int
	ImageHeight int
}

// A grid to represent the current state

func makeWorld(IoInput chan uint8, height int, width int) [][]uint8 {
	world := make([][]uint8, height)
	for row := 0; row < height; row++ {
		world[row] = make([]uint8, width)
		for cell := 0; cell < width; cell++ {
			world[row][cell] = <-IoInput
		}
	}
	return world
}

// Run starts the processing of Game of Life. It should initialise channels and goroutines.
func Run(p Params, events chan<- Event, keyPresses <-chan rune) {

	IoCommand := make(chan ioCommand)
	IoFilename := make(chan string)
	IoIdle := make(chan bool)
	IoOutput := make(chan uint8)
	IoInput := make(chan uint8)

	ioChannels := ioChannels{
		command:  IoCommand,
		idle:     IoIdle,
		filename: IoFilename,
		output:   IoOutput,
		input:    IoInput,
	}
	go startIo(p, ioChannels)

	// Input data
	IoCommand <- ioInput
	imageName := fmt.Sprintf("%dx%d", p.ImageHeight, p.ImageWidth)
	IoFilename <- imageName

	world := makeWorld(IoInput, p.ImageWidth, p.ImageHeight)

	clientChannels := clientChannels{
		events,
		IoCommand,
		IoIdle,
		IoFilename,
		IoInput,
		IoOutput,
		keyPresses,
	}
	client := Client{}

	// Establish contact with the server
	srvrAddr := "100.84.31.131:8030"
	server, err := rpc.Dial("tcp", srvrAddr)
	handleError(err)

	// err = server.Call(, args, reply)
	// handleError(err)

	r := stubs.Parameters{p.ImageHeight, p.ImageWidth, p.Turns, p.Threads}

	// Request the initial world and all its parameters
	req := stubs.Request{World: world, Param: r}

	// Respond with the final world and all its parameters
	res := new(stubs.Response)

	err = server.Call(stubs.BeginWorld, req, res)
	handleError(err)

	go client.clientRun(r, clientChannels, server)
}
