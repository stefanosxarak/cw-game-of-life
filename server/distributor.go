package main

import (
	"uk.ac.bris.cs/gameoflife/util"
)

type Data struct {
	world       [][]uint8
	newWorld    [][]uint8
	threads     int
	imageWidth  int
	imageHeight int
	turns       int
	quit        bool
}

//error struct
type errorString struct {
	s string
}

const alive = 255
const dead = 0

func mod(x, m int) int {
	return (x + m) % m
}

//calculateNeighbors takes the current state of the world and completes one evolution of the world. It then returns the result
func (d *Data) calculateNeighbours(x, y int, world [][]uint8) int {
	neighbours := 0
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			if i != 0 || j != 0 {
				if world[mod(y+i, d.imageHeight)][mod(x+j, d.imageWidth)] == alive {
					neighbours++
				}
			}
		}
	}
	return neighbours
}

//progress to next state and update CellFlipped event
func (d *Data) calculateNextState(world [][]uint8) [][]uint8 {
	newWorld := make([][]uint8, d.imageHeight)
	for i := range newWorld {
		newWorld[i] = make([]uint8, d.imageWidth)
	}
	for y := 0; y < d.imageHeight; y++ {
		for x := 0; x < d.imageWidth; x++ {
			neighbours := d.calculateNeighbours(x, y, world)
			if world[y][x] == alive {
				if neighbours == 2 || neighbours == 3 {
					newWorld[y][x] = alive
				} else {
					newWorld[y][x] = dead
				}
			} else {
				if neighbours == 3 {
					newWorld[y][x] = alive
				} else {
					newWorld[y][x] = dead
				}
			}
		}
	}
	return newWorld
}

// returns a slice of the alive cells in prevWorld
// Could not add parameters as server gets complicated
func (d *Data) calculateAliveCells() []util.Cell {
	aliveCells := make([]util.Cell, 0)
	for row := range d.world {
		for col := range d.world[row] {
			if d.world[row][col] == alive {
				aliveCells = append(aliveCells, util.Cell{X: col, Y: row})
			}
		}
	}
	return aliveCells
}

//A 2D slice to store the updated world.
func (d *Data) makeNewWorld(height int, width int) {
	d.newWorld = make([][]uint8, height)
	for row := 0; row < height; row++ {
		d.newWorld[row] = make([]uint8, width)
	}
}

// distributor divides the work between workers and interacts with other goroutines.
func (d *Data) distributor() {

	//Initialization
	d.makeNewWorld(d.imageHeight, d.imageWidth)

	//Game of Life.
	for d.turns = 0; d.turns < d.turns && d.quit == false; d.turns++ {

		//we add the newly updated world to the grid we had made
		d.newWorld = d.calculateNextState(d.world)
		d.world = d.newWorld
	}
}
