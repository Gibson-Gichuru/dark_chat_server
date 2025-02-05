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
	MessageType uint8 = iota + 1
	HeartBeat
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

	encoded := base64.StdEncoding.EncodeToString(m.Byte())
	o, err := w.Write([]byte(encoded))

	return int64(o), err

}

// ReadFrom implements the io.ReaderFrom interface.
// It reads a message from the reader with a message type header.
// The message is assigned to the Message receiver.
// It returns the number of bytes read and any error encountered.
// If the size of the message exceeds MaxPayloadsize, it returns an error.

func (m *Message) ReadFrom(r io.Reader) (int64, error) {

	var size uint32

	err := binary.Read(r, binary.BigEndian, &size)

	if err != nil {
		return 0, err
	}

	buf := make([]byte, size)

	n, err := r.Read(buf)

	if err != nil {
		return int64(n), err
	}

	decoded, err := base64.StdEncoding.DecodeString(string(buf))

	if err != nil {
		return int64(n), err
	}

	*m = Message(decoded)

	return int64(n), nil
}

// Decode reads a message from the reader with a message type header, and
// decodes it into a Payload. It returns the Payload and any error encountered.
// If the message type is unknown, it returns an error.
func Decode(r io.Reader) (Payload, error) {

	var typ uint8

	err := binary.Read(r, binary.BigEndian, &typ)

	if err != nil {
		return nil, err
	}

	switch typ {
	case HeartBeat:
		return nil, nil

	case MessageType:
		return decodeMessageType(r)
	default:
		return nil, ErrorUnknownType
	}
}

func decodeMessageType(r io.Reader) (Payload, error) {

	var size uint8
	var headers PayloadHeaders
	var payload = new(Message)

	err := binary.Read(r, binary.BigEndian, &size)

	if err != nil {
		return nil, err
	}

	headerBuf := make([]byte, size)

	_, err = r.Read(headerBuf)

	if err != nil {
		return nil, err
	}

	decoded, err := base64.StdEncoding.DecodeString(string(headerBuf))

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(decoded, &headers)

	if err != nil {
		return nil, err
	}

	if headers.Size > MaxPayloadsize {
		return nil, ErrorMaxPayloadSize
	}

	payloadSize := headers.Size

	payloadSizeBuf := new(bytes.Buffer)

	err = binary.Write(payloadSizeBuf, binary.BigEndian, payloadSize)

	if err != nil {
		return nil, err
	}

	_, err = payload.ReadFrom(
		io.MultiReader(payloadSizeBuf, r),
	)

	if err != nil {
		return nil, err
	}

	return payload, nil
}

func Encode(w io.Writer, payload Payload, payloadType uint8) (int64, error) {

	switch payloadType {
	case HeartBeat:
		err := binary.Write(w, binary.BigEndian, HeartBeat)

		if err != nil {
			return 0, err
		}

		return 1, nil
	case MessageType:

		return encodeMessageType(w, payload)

	default:
		return 0, ErrorUnknownType
	}

}

func encodeMessageType(w io.Writer, payload Payload) (int64, error) {

	err := binary.Write(w, binary.BigEndian, MessageType)

	if err != nil {
		return 0, err
	}

	var n int64 = 1

	payloadHeaders := PayloadHeaders{
		Size: uint32(len(
			base64.StdEncoding.EncodeToString(payload.Byte()),
		)),
	}

	payloadHeadersJson, err := json.Marshal(payloadHeaders)

	if err != nil {
		return 1, err
	}

	encoded := base64.StdEncoding.EncodeToString(payloadHeadersJson)

	err = binary.Write(w, binary.BigEndian, uint8(len(encoded)))

	if err != nil {
		return n, err
	}

	_, err = w.Write([]byte(encoded))

	if err != nil {
		return n, nil
	}

	n += int64(len(encoded))

	o, err := payload.WriteTo(w)

	return n + o, err

}
