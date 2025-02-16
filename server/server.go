package server

import (
	"context"
	"darkchat/database"
	"darkchat/monitor"
	"darkchat/pinger"
	"darkchat/protocol"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/google/uuid"
)

const DEFAULTPINGINTERVAL = 30 * time.Second

const (
	REXTENTION uint8 = iota + 1
	WEXTENTION
	RWEXTENTION
)

var monitorLogger = monitor.New("server.log")

type ConnectionBuilder struct {
	ConnectionType string
	Address        string
	Port           string
}

type Client struct {
	chatId     string
	connection net.Conn
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
		monitorLogger.Fatal(err.Error())
		os.Exit(1)
	}

	monitorLogger.Info(fmt.Sprintf("Listening on %s", builder.Addressbuilder()))

	defer server.Close()

	for {
		conn, err := server.Accept()
		if err != nil {
			monitorLogger.Error(err.Error())
			continue
		}

		client := Client{
			connection: conn,
			chatId:     uuid.NewString(),
		}

		monitorLogger.Info(fmt.Sprintf("Accepted connection from %s", conn.RemoteAddr().String()))

		go handleClientConnection(client)

	}
}

// handleClientConnection manages the lifecycle of a client's connection. It registers the client's chat ID
// with the database, starts a pinger to send periodic "PING" messages, and listens for incoming messages
// from the client. Incoming messages are decoded and posted to the chat stream. Outgoing messages from the
// chat stream are sent to the client. The function handles connection cleanup and error logging.

func handleClientConnection(client Client) {
	ctx, cancel := context.WithCancel(context.Background())
	var streamingChanel = make(chan protocol.Payload, 20)
	var clientStreamSubChannel = make(chan string, 1)
	clientStreamSubChannel <- client.chatId

	defer func() {
		cancel()
		client.connection.Close()
		close(clientStreamSubChannel)

		err := database.DeleteClientChat(client.chatId)

		if err != nil {
			monitorLogger.Error(err.Error())
		}

	}()

	dbErr := database.RegisterClientChat(client.chatId)

	if dbErr != nil {
		monitorLogger.Error(dbErr.Error())
		return
	}
	resetTimer := make(chan time.Duration, 1)
	resetTimer <- time.Second

	go pinger.Ping(ctx, client.connection, resetTimer)

	if err := extendDeadline(client.connection, DEFAULTPINGINTERVAL, RWEXTENTION); err != nil {
		return
	}

	go database.StreamChat(streamingChanel, clientStreamSubChannel, client.chatId)

	go func() {
		for message := range streamingChanel {
			err := writeToClient(client, message, protocol.MessageType)
			if err != nil {
				monitorLogger.Error(err.Error())
			}
		}
	}()

	for {

		message, err := protocol.Decode(client.connection)

		if err != nil {
			monitorLogger.Error(err.Error())
			return
		}
		resetTimer <- 0

		if err := extendDeadline(client.connection, DEFAULTPINGINTERVAL, WEXTENTION); err != nil {
			return
		}

		var m protocol.Message

		err = json.Unmarshal(message.Byte(), &m)

		if err != nil {
			monitorLogger.Error(err.Error())
			return
		}

		if !database.CheckChatExists(m.To) {

			err := protocol.Error_("chat does not exist")
			if clientErr := writeToClient(client, &err, protocol.Error); clientErr != nil {
				return
			}
			continue
		}

		database.PostToChat(m.String(), m.To)

	}
}

// writeToClient writes the given message to the client connection, with the
// given message type, and resets the connection deadline to the default ping
// interval. It returns an error if there was an error writing to the client or
// extending the deadline.
func writeToClient(client Client, message protocol.Payload, messageType uint8) error {
	_, err := protocol.Encode(
		client.connection,
		message,
		messageType,
	)
	if err != nil {
		return err
	}

	if internalError := extendDeadline(client.connection, DEFAULTPINGINTERVAL, REXTENTION); internalError != nil {
		return internalError
	}
	return nil
}

// extendDeadline sets the deadline for the given connection to the current time plus the given duration.
// If an error occurs while setting the deadline, the error is logged and returned.
// Otherwise, the function returns nil.
func extendDeadline(conn net.Conn, duration time.Duration, extentionType uint8) error {

	switch extentionType {
	case REXTENTION:
		err := conn.SetReadDeadline(time.Now().Add(duration))
		if err != nil {
			monitorLogger.Error(err.Error())
			return err
		}
	case WEXTENTION:
		err := conn.SetWriteDeadline(time.Now().Add(duration))
		if err != nil {
			monitorLogger.Error(err.Error())
			return err
		}

	case RWEXTENTION:
		err := conn.SetDeadline(time.Now().Add(duration))
		if err != nil {
			monitorLogger.Error(err.Error())
			return err
		}
	}
	return nil
}
