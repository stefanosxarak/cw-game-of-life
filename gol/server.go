package gol

import (
	"flag"
	"fmt"
	"net"
	"net/rpc"
)


//TODO:
// Sends a AliveCellsCount event to Client every 2 seconds

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {

	// parse compiler flags
	fmt.Println("Launching server...")
	fmt.Println("Listen on port")
	portPtr := flag.String("this", "8030", "Port to listen on")
	flag.Parse()

	// register the interface
	// rpc.Register(new(Server))
	// listen for calls
	active := true
	for active {
		ln, err := net.Listen("tcp", *portPtr)
		handleError(err)
		defer ln.Close()
		// accept a listener
		go rpc.Accept(ln)
	}
}
