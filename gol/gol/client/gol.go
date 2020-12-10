package gol

import (
	"uk.ac.bris.cs/gameoflife/gol/stubs"
	"fmt"
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

	// Input data
	ioCommand <- ioInput
	imageName := fmt.Sprintf("%dx%d", p.ImageHeight, p.ImageWidth)
	ioFilename <- imageName

	args := stubs.StartArgs{
		Turns:   p.Turns,
		Threads: p.Threads,
		Height:  p.ImageHeight,
		Width:   p.ImageWidth,
	}

	clientChannels := ClientChans{
		events,
		IoCommand,
		IoIdle,
		IoFilename,
		IoInput,
		IoOutput,
		keyPresses,
	}
	client := Client{}

	go clientRun(p, clientChannels, server)
	// Make sure that the Io has finished any output before exiting.
	// events <- ImageOutputComplete{turn, imageName}
}