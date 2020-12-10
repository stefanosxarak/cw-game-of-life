package gol

import (
	"flag"
	"fmt"
	"net"
	"net/rpc"
)

// Visit:
// https://golang.org/pkg/net/rpc/ for info about how rpc works
// https://golang.org/src/runtime/stubs.go for info about how stubs work or check the DS lab

type ServerInterface struct {
}

//TODO:
// Implement ServerInterface
// Sends a AliveCellsCount event to Client every 2 seconds
// Kill server
// Pause/unpause server

// func (s *Server) AliveCells(args stubs.Default, reply *stubs.Alive) error {

// 	reply.Num = len(s.distributor.CalculateAliveCells())
// 	reply.Turn = s.distributor.currentTurn

// 	return nil
// }

func main() {
	// parse compiler flags
	fmt.Println("Launching server...")
	fmt.Println("Listen on port")
	portPtr := flag.String("this", "8030", "Port to listen on")
	flag.Parse()

	// register the interface
	server := new(ServerInterface)
	rpc.Register(server)

	// Awaiting connection
	ln, err := net.Listen("tcp", *portPtr)
	if err != nil {
		panic(err)
	}
	defer ln.Close()
	rpc.Accept(ln)
}
