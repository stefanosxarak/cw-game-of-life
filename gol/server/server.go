package main

import (
	"flag"
	"fmt"
	"net"
	"net/rpc"

	"uk.ac.bris.cs/gameoflife/stubs"
)

// Visit:
// https://golang.org/pkg/net/rpc/ for info about how rpc works
// https://golang.org/src/runtime/stubs.go for info about how stubs work or check the DS lab

type ServerInterface struct {
	// distributor Distributor
}

//TODO:
// Implement ServerInterface
// Sends a AliveCellsCount event to Client every 2 seconds
// Kill server

// func (s *ServerInterface) AliveCells(args stubs.Default, reply *stubs.Alive) error {

// 	reply.Num = len(s.distributor.CalculateAliveCells())
// 	reply.Turn = s.distributor.currentTurn

// 	return nil
// }

// beginWorlds starts processing worlds
// func (s *ServerInterface) beginWorlds(args stubs.StartArgs, reply *stubs.Default) error {

// 	s.distributor = Distributor{
// 		currentTurn: 0,
// 		numOfTurns:  args.Turns,
// 		threads:     args.Threads,
// 		imageWidth:  args.Width,
// 		imageHeight: args.Height,
// 		prevWorld:   WorldSlice,
// 		paused:      make(chan bool),
// 	}
// 	go s.distributor.distributor(p)
// 	return nil
// }

// Sends a proccessed world from Server
func (s *ServerInterface) worldFromServer(args stubs.Default, reply *stubs.Request) error {
	// TODO: Take the correct data from distributor 
	reply.World = s.distributor.prevWorld
	reply.Param.Turns = s.distributor.currentTurn
	reply.Param.ImageHeight = s.distributor.imageHeight
	reply.Param.ImageWidth = s.distributor.imageWidth

	return nil
}

func main() {
	// parse compiler flags
	fmt.Println("Launching server...")
	fmt.Println("Listen on port")
	portPtr := flag.String("this", "8030", "Port to listen on")
	flag.Parse()

	// register the Server
	server := new(ServerInterface)
	rpc.Register(server)

	// Awaiting connection
	ln, err := net.Listen("tcp",": " +*portPtr)
	if err != nil {
		panic(err)
	}
	defer ln.Close()
	rpc.Accept(ln)
}
