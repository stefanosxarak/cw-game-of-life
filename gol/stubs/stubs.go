package stubs

// StartGoL starts a Game of Life simulation on the server.
// args = StartArgs, reply = Default
var StartGoL = "Server.StartGoL"

// GetWorld reutnrs the latest nXn world form server where m is integer greater than zero.
// args = Default, reply = WorldGeneric
var GetWorld = "Server.GetWorld"

// Connect establishes a connection between client and server.
// args = Default, reply = Status
var Connect = "Server.Connect"

// AliveCells gives a report on the number of alive cells and the current turn.
// args = Default, reply = Alive
var AliveCells = "Server.AliveCells"

// Pause stops the server until further notice.
// args = Default, reply = Turn
var Pause = "Server.Pause"

// Kill shuts down the server.
// args = Default, reply = Turn
var Kill = "Server.Kill"

// StartArgs provides the initial conditions for GoL
type StartArgs struct {
	Turns   int
	Threads int
	Height  int
	Width   int
	World   [][]uint8
}

// Default args/reply for all methods
type Default struct{}

//World contains world encoded in [][]uint8 format
type World struct {
	World  [][]uint8
	Height int
	Width  int
	Turn   int
}

// Turn contains the current turn
type Turn struct {
	Turn int
}

// Alive contains the number of alive cells
type Alive struct {
	Num  int
	Turn int
}

// Status contains details about the engine's current simulation.
// if Running, it will include the state (the following variables)
type Status struct {
	Running     bool
	Width       int
	Height      int
	CurrentTurn int
	NumOfTurns  int
}

// Done contains a boolean that represents whether or not the server has finished processing
type Done struct {
	Done bool
}
