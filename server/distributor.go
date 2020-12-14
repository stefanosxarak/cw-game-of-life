package main

import (
	"uk.ac.bris.cs/gameoflife/stubs"
)

// type distributorChannels struct {
// 	events     chan<- Event
// 	ioCommand  chan<- ioCommand
// 	ioIdle     <-chan bool
// 	ioFilename chan<- string
// 	ioOutput   chan<- uint8
// 	ioInput    <-chan uint8
// 	keyPresses <-chan rune
// }
type Data struct {
	world       [][]uint8
	newWorld    [][]uint8
	threads     int
	imageWidth  int
	imageHeight int
	turns       int
}

//error struct
type errorString struct {
	s string
}



func mod(x, m int) int {
	return (x + m) % m
}

//calculateNeighbors takes the current state of the world and completes one evolution of the world. It then returns the result
func calculateNeighbours(p stubs.Parameters, x, y int, world [][]byte) int {
	neighbours := 0
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			if i != 0 || j != 0 {
				if world[mod(y+i, p.ImageHeight)][mod(x+j, p.ImageWidth)] == alive {
					neighbours++
				}
			}
		}
	}
	return neighbours
}

// //progress to next state and update CellFlipped event
// func calculateNextState(p stubs.Parameters, turn int, c distributorChannels, world [][]byte) [][]byte {
// 	newWorld := make([][]byte, p.ImageHeight)
// 	for i := range newWorld {
// 		newWorld[i] = make([]byte, p.ImageWidth)
// 	}
// 	for y := 0; y < p.ImageHeight; y++ {
// 		for x := 0; x < p.ImageWidth; x++ {
// 			neighbours := calculateNeighbours(p, x, y, world)
// 			if world[y][x] == alive {
// 				if neighbours == 2 || neighbours == 3 {
// 					newWorld[y][x] = alive
// 				} else {
// 					newWorld[y][x] = dead
// 					c.events <- CellFlipped{turn, util.Cell{X: x, Y: y}}
// 				}
// 			} else {
// 				if neighbours == 3 {
// 					newWorld[y][x] = alive
// 					c.events <- CellFlipped{turn, util.Cell{X: x, Y: y}}
// 				} else {
// 					newWorld[y][x] = dead
// 				}
// 			}
// 		}
// 	}
// 	return newWorld
// }

//A 2D slice to store the updated world.
func (d *Data) makeNewWorld(height int, width int) {
	d.newWorld = make([][]uint8, height)
	for row := 0; row < height; row++ {
		d.newWorld[row] = make([]uint8, width)
	}
}

// distributor divides the work between workers and interacts with other goroutines.
func (d *Data) distributor(p stubs.Parameters) {

	//Initialization
	d.makeNewWorld(p.ImageHeight, p.ImageWidth)

	//Game of Life.
	var turn int
	for turn = 0; turn < p.Turns; turn++ {

		//we add the newly updated world to the grid we had made
		d.world = d.newWorld
		d.makeNewWorld(p.ImageHeight, p.ImageWidth)

	}

}
