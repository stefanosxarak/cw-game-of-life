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


// Interface
type Server struct {
	data Data
}

//TODO:
// fix bug with beginworlds
// Sends a AliveCellsCount event to Client every 2 seconds


// beginWorlds starts processing worlds
func (s *Server) beginWorld(args stubs.Request, reply *stubs.Default) error {
	s.data = Data{
		world:       args.World,
		imageHeight: args.Param.ImageHeight,
		imageWidth:  args.Param.ImageWidth,
		turns:       args.Param.Turns,
		threads:     args.Param.Threads,
	}
	go s.data.distributor()
	return nil
}

// Kill shuts down the server
func (s *Server) Kill(args stubs.Default, reply *stubs.Parameters) error {
	s.data.quit = true
	reply.Turns = s.data.turns
	return nil
}

// Sends a proccessed world from Server
func (s *Server) worldFromServer(args stubs.Default, reply *stubs.Request) error {
	reply.World = s.data.world
	reply.Param.Turns = s.data.turns
	reply.Param.ImageHeight = s.data.imageHeight
	reply.Param.ImageWidth = s.data.imageWidth

	return nil
}

func main() {
	// parse compiler flags
	fmt.Println("Launching server...")
	fmt.Println("Listen on port")
	portPtr := flag.String("this", "8030", "Port to listen on")
	flag.Parse()

	// register the Server
	rpc.Register(new(Server))

	// Awaiting connection
	ln, err := net.Listen("tcp", ": " +*portPtr)
	if err != nil {
		panic(err)
	}
	fmt.Print("Server is active...")
	defer ln.Close()
	rpc.Accept(ln)
}
