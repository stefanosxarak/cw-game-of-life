package stubs

type Response struct {
	W [][]byte
}

type Request struct {
	W     [][]byte
	Param Parameters
}

type Parameters struct {
	Turns       int
	Threads     int
	ImageWidth  int
	ImageHeight int
}
