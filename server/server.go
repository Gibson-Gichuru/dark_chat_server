package server

import (
	"darkchat/protocol"
	"fmt"
	"log"
	"net"
)

type ConnectionBuilder struct {
	ConnectionType string
	Address        string
	Port           string
}

// Addressbuilder constructs and returns a string representing the full network address
// by combining the Address and Port fields of the ConnectionBuilder.

func (c ConnectionBuilder) Addressbuilder() string {
	return fmt.Sprintf("%s:%s", c.Address, c.Port)
}

// ServerStart starts a server listening on the address specified by the
// ConnectionBuilder, and accepts incoming connections. Each connection is
// handled in a separate goroutine by calling handleClientConnection. If an
// error occurs while accepting a connection, the error is logged and the
// function continues. The function does not return until an error occurs while
// listening.
func ServerStart(builder ConnectionBuilder) {

	server, err := net.Listen(builder.ConnectionType, builder.Addressbuilder())

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Listening on %s", server.Addr())

	defer server.Close()

	for {
		conn, err := server.Accept()

		if err != nil {
			log.Println(err)
			continue
		}

		go handleClientConnection(conn)

	}
}

// handleClientConnection is a helper function that is called in a separate
// goroutine for each incoming connection. It reads messages from the
// connection, decodes them, and logs them to the console. If an error occurs
// while reading from the connection, the error is logged and the function
// continues to the next iteration. The function does not return until the
// connection is closed.
func handleClientConnection(conn net.Conn) {
	defer conn.Close()

	for {
		payload, err := protocol.Decode(conn)

		if err != nil {
			log.Println(err)
			continue
		}
		fmt.Printf("Received: %s\n", payload)
	}
}
