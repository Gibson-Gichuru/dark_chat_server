package server

import (
	"context"
	"darkchat/monitor"
	"darkchat/pinger"
	"darkchat/protocol"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

type Server struct {
	*ConnectionBuilder
	*monitor.Monitor
}

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
func ServerStart(builder Server) {

	server, err := net.Listen(builder.ConnectionType, builder.Addressbuilder())

	if err != nil {
		builder.Fatal(err.Error())
		os.Exit(1)
	}

	builder.Info(fmt.Sprintf("Listening on %s", builder.Addressbuilder()))

	defer server.Close()

	for {
		conn, err := server.Accept()

		if err != nil {
			builder.Error(err.Error())
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
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		conn.Close()
	}()

	resetTimer := make(chan time.Duration, 1)
	resetTimer <- time.Second

	go pinger.Ping(ctx, conn, resetTimer)

	err := conn.SetDeadline(time.Now().Add(10 * time.Second))

	if err != nil {
		log.Println(err)
		return
	}
	for {
		payload, err := protocol.Decode(conn)

		if err != nil {
			log.Println(err)
			return
		}
		resetTimer <- 0

		err = conn.SetDeadline(time.Now().Add(10 * time.Second))

		if err != nil {
			log.Println(err)
			return
		}
		fmt.Printf("Received: %s\n", payload)
	}
}
