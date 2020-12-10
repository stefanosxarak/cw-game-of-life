package gol

import (
	"net/rpc"

	"uk.ac.bris.cs/gameoflife/gol/stubs"
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

	// go distributor(p, distributorChannels)

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
	// defer server.Close()
	// err = server.Call(, args, reply)
	handleError(err)

	def := new(stubs.Default)
	status := new(stubs.Status)
	server.Call(stubs.Connect, def, status)

	args := stubs.StartArgs{
		Turns:   p.Turns,
		Threads: p.Threads,
		Height:  p.ImageHeight,
		Width:   p.ImageWidth,
	}

	clientChans := clientChans{
		events,
		IoCommand,
		IoIdle,
		IoFilename,
		IoInput,
		IoOutput,
		keyPresses,
	}

	go clientRun(p, clientChans, server)	
}
