package gol

import (
	"fmt"
	"os"
)

type ClientChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioInput    <-chan uint8
	ioOutput   chan<- uint8
	keyPresses <-chan rune
}

func handleError(err error) {
	fmt.Println(err)
	os.Exit(1)
	// TODO: all
	// Deal with an error event.
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
