package stubs

var worldFromServer = "ServerInterface.worldFromServer"

// Empty struct that acts as default args
type Default struct{}

type Response struct {
	World [][]uint8
}

type Request struct {
	World     [][]uint8
	Param Parameters
}

// Could have integrated this struct with the Request but no time!
type Parameters struct {
	ImageHeight int
	ImageWidth  int
	Turns       int
	// Threads     int
}