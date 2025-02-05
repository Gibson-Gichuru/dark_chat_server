package protocol

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

const (
	MessageType    uint8  = iota + 1
	MaxPayloadsize uint32 = 10 << 20
)

var ErrorMaxPayloadSize = errors.New("max payload size exceeded")
var ErrorEmptyHeaders = errors.New("empty headers")
var ErrorUnknownType = errors.New("unknown message type")

type Payload interface {
	fmt.Stringer
	io.WriterTo
	io.ReaderFrom
	Byte() []byte
}

type PayloadHeaders struct {
	Size     uint32
	Type     uint8
	Encoding string
}

type Message string

func (m Message) String() string { return string(m) }
func (m Message) Byte() []byte   { return []byte(m) }

// WriteTo implements the io.WriterTo interface.
// It writes a message to the writer with a message type header, and returns the number of bytes written and any error encountered.
func (m Message) WriteTo(w io.Writer) (int64, error) {

	o, err := w.Write([]byte(m))

	return int64(o), err

}

// ReadFrom implements the io.ReaderFrom interface.
// It reads a message from the reader with a message type header.
// The message is assigned to the Message receiver.
// It returns the number of bytes read and any error encountered.
// If the size of the message exceeds MaxPayloadsize, it returns an error.

func (m *Message) ReadFrom(r io.Reader) (int64, error) {

	buf := new(bytes.Buffer)

	n, err := buf.ReadFrom(r)

	if err != nil {
		return 0, err
	}

	*m = Message(buf.String())

	return n + int64(n), err
}

// Decode reads a message from the reader with a message type header, and
// decodes it into a Payload. It returns the Payload and any error encountered.
// If the message type is unknown, it returns an error.
func Decode(r io.Reader) (Payload, error) {

	var headers PayloadHeaders

	var payloadHeadersLen uint8

	var payload Payload

	headersBuf := new(bytes.Buffer)

	err := binary.Read(r, binary.BigEndian, &payloadHeadersLen)

	if err != nil {
		return nil, err
	}

	io.CopyN(headersBuf, r, int64(payloadHeadersLen))

	decoded, err := base64.StdEncoding.DecodeString(headersBuf.String())

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(decoded, &headers)

	if err != nil {
		return nil, ErrorEmptyHeaders
	}

	switch headers.Type {
	case MessageType:
		payload = new(Message)
	default:
		return nil, ErrorUnknownType
	}

	_, err = payload.ReadFrom(r)

	return payload, err
}

func Encode(w io.Writer, payload Payload, payloadType uint8) (int64, error) {

	msgHeaders := PayloadHeaders{Type: payloadType, Size: uint32(len(payload.Byte()))}

	msgHeadersStr, err := json.Marshal(msgHeaders)

	if err != nil {
		return 0, err
	}
	encoded := base64.StdEncoding.EncodeToString(msgHeadersStr)

	err = binary.Write(w, binary.BigEndian, uint8(len(encoded)))

	if err != nil {
		return 0, err
	}

	var n int64 = 1

	io.CopyN(w, bytes.NewReader([]byte(encoded)), int64(len(encoded)))

	n += 8

	o, err := payload.WriteTo(w)

	return n + o, err

}
