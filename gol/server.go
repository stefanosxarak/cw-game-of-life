package gol

import (
	"flag"
	"net"
	"net/rpc"
)

func main() {
	// parse compiler flags
	port := flag.String("this", "8030", "Port for this service to listen on")
	flag.Parse()
	// register the interface
	// rpc.Register(new(Server))
	// listen for calls
	active := true
	for active {
		listener, err := net.Listen("tcp", ":"+*port)
		if err != nil {
			panic(err)
		}
		defer listener.Close()
		// accept a listener
		go rpc.Accept(listener)
	}
}
