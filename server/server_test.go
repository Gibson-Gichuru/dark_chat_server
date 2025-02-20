package server

import (
	"context"
	"darkchat/protocol"
	"net"
	"testing"
)

func TestClientServerCom(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	connectionBuilder := ConnectionBuilder{ConnectionType: "tcp", Address: "localhost", Port: "8090"}

	go ServerStart(ctx, connectionBuilder)

	con, err := net.Dial("tcp", "localhost:8090")

	if err != nil {
		t.Fatal(err)
	}

	defer con.Close()

	message := protocol.Message{
		Message: "Hello, world",
		From:    "",
		To:      "",
	}

	_, err = protocol.Encode(con, &message, protocol.MessageType)

	if err != nil {
		t.Fatal(err)
	}


	p, err := protocol.Decode(con)

	if err != nil {
		t.Fatal(err)
	}

	switch p.(type) {
	case *protocol.Error_:
		t.Log("Message decoded successfully")
	default:
		t.Fatal("Expected Message got ", p)
	}
}
