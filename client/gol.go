package gol

import (
	"uk.ac.bris.cs/gameoflife/stubs"
	"net/rpc"
)

// Params provides the details of how to run the Game of Life and which image to load.
type Params struct {
	Turns       int
	Threads     int
	ImageWidth  int
	ImageHeight int
}

// Run starts the processing of Game of Life. It should initialise channels and goroutines.
func Run(p Params, events chan<- Event, keyPresses <-chan rune) {

	ioCommand := make(chan ioCommand)
	ioFilename := make(chan string)
	ioIdle := make(chan bool)
	ioOutput := make(chan uint8)
	ioInput := make(chan uint8)

	// distributorChannels := distributorChannels{
	// 	events,
	// 	ioCommand,
	// 	ioIdle,
	// 	ioFilename,
	// 	ioInput,
	// 	ioOutput,
	// 	keyPresses,
	// }

	ioChannels := ioChannels{
		command:  ioCommand,
		idle:     ioIdle,
		filename: ioFilename,
		output:   ioOutput,
		input:    ioInput,
	}
	go startIo(p, ioChannels)

	// Establish contact with the server
	srvrAddr := "localhost:8030"
	server, err := rpc.Dial("tcp", srvrAddr)
	handleError(err)
	defer server.Close()
	// err = server.Call(, args, reply)
	handleError(err)

	r := stubs.Parameters{p.ImageHeight, p.ImageWidth, p.Turns}
	// Request the initial world and all its parameters
	req := new(stubs.Request)

	// Respond with the final world and all its parameters
	res := new(stubs.Response)
	err = server.Call(stubs.NextState, request, response)
	handleError(err)


	clientChannels := clientChans{
		events,
		IoCommand,
		IoIdle,
		IoFilename,
		IoInput,
		IoOutput,
		keyPresses,
	}
	// go clientRun(p, clientChannels, server)
}
