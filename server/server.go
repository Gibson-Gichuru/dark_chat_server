package server

import (
	"fmt"
	"io"
	"log"
	"net"
)

type ConnectionBuilder struct {
	ConnectionType string
	Address        string
	Port           string
}

func (c ConnectionBuilder) Addressbuilder() string {
	return fmt.Sprintf("%s:%s", c.Address, c.Port)
}

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
func handleClientConnection(conn net.Conn) {
	defer conn.Close()

	for {
		buffer := make([]byte, 1024)
		_, err := conn.Read(buffer)

		if err != nil {
			if err != io.EOF {
				log.Println(err)
			}
			return
		}
	}
}
