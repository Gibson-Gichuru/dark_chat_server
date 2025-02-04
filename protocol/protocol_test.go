package protocol

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"testing"
)

// TestPayloadEncoding tests that a message can be encoded with a message type header
// correctly. It creates a message, encodes it, and checks that the decoded headers
// are equal to the original message type.
func TestPayloadEncoding(t *testing.T) {
	message := Message("This some cool communication protocol")

	buf := new(bytes.Buffer)

	_, err := Encode(buf, &message, MessageType)

	if err != nil {
		t.Error(err)
	}

	// check that the headers where encoded correctly

	var payloadHeaderLen uint8

	err = binary.Read(buf, binary.BigEndian, &payloadHeaderLen)

	if err != nil {
		t.Errorf("Error reading payload header length: %s", err)
	}
	headerBuf := new(bytes.Buffer)

	io.CopyN(headerBuf, buf, int64(payloadHeaderLen))

	decoded, err := base64.StdEncoding.DecodeString(headerBuf.String())

	if err != nil {
		t.Errorf("Error decoding headers: %s", err)
	}

	var headers PayloadHeaders

	err = json.Unmarshal(decoded, &headers)

	if err != nil {
		t.Errorf("Error unmarshalling headers: %s", err)
	}

	if headers.Type != MessageType {
		t.Errorf("Expected message type %d, got %d", MessageType, headers.Type)
	}

}

// TestPayloadDecode tests that a message can be encoded and then decoded
// correctly. It creates a message, encodes it, decodes it, and checks that the
// decoded message is equal to the original message.
func TestPayloadDecode(t *testing.T) {

	message := Message("This some cool communication protocol")

	buf := new(bytes.Buffer)

	_, err := Encode(buf, &message, MessageType)

	if err != nil {
		t.Errorf("Error encoding message: %s", err)
	}

	payload, err := Decode(buf)

	if err != nil {
		t.Errorf("Error decoding message: %s", err)
	}

	if payload.String() != message.String() {
		t.Errorf("Expected message %s, got %s", message, payload)
	}

	fmt.Println(payload)

}
