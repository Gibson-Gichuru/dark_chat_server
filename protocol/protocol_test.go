package protocol

import (
	"bytes"
	"testing"
)

// TestPayloadEncoding tests that a message can be encoded with a message type header
// correctly. It creates a message, encodes it, and checks that the decoded headers
// are equal to the original message type.
func TestHeartBeatPayloadEncoding(t *testing.T) {

	buf := new(bytes.Buffer)

	_, err := Encode(buf, nil, HeartBeat)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestHearBeatPayloadDecoding(t *testing.T) {
	buf := new(bytes.Buffer)

	_, err := Encode(buf, nil, HeartBeat)
	if err != nil {
		t.Errorf("Expected no Encoding error, got %v", err)
	}

	m, err := Decode(buf)
	if err != nil {
		t.Errorf("Expected no Decoding error, got %v", err)
	}

	if m != nil {
		t.Errorf("Expected nil, got %v", m)
	}

}

// TestPayloadDecode tests that a message can be encoded and then decoded
// correctly. It creates a message, encodes it, decodes it, and checks that the
// decoded message is equal to the original message.
func TestMessagePayloadEncoding(t *testing.T) {
	message := Message("Hello, world!")

	buf := new(bytes.Buffer)

	_, err := Encode(buf, &message, MessageType)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestMessagePayloadDecoding(t *testing.T) {
	message := Message("Hello, world")

	buf := new(bytes.Buffer)

	_, err := Encode(buf, &message, MessageType)

	if err != nil {
		t.Errorf("Expected no Encoding error, got %v", err)
	}

	m, err := Decode(buf)

	if err != nil {
		t.Errorf("Expected not Decoding error got: %v", err)
	}

	if message.String() != m.String() {
		t.Errorf("Expected m to be : %s got : %s", message.String(), m.String())
	}
}
