package main

import (
	"darkchat/server"
	"flag"
	"fmt"
)

var (
	address = flag.String("address", "127.0.0.1", "The address to bind to")
	port    = flag.String("port", "8080", "The port to bind to")
)

func init() {
	flag.Usage = func() {
		fmt.Printf("Usage: %s -address=127.0.0.1 -port=8080\n", flag.CommandLine.Name())
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	s := server.ConnectionBuilder{
		ConnectionType: "tcp",
		Address:        *address,
		Port:           *port,
	}
	server.ServerStart(s)
}
