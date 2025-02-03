package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	MessageType    uint8  = iota + 1
	MaxPayloadsize uint32 = 10 << 20
)

var ErrorMaxPayloadSize = errors.New("max payload size exceeded")

type Payload interface {
	fmt.Stringer
	io.WriterTo
	io.ReaderFrom
	Byte() []byte
}

type Message string

func (m Message) String() string { return string(m) }
func (m Message) Byte() []byte   { return []byte(m) }

func (m Message) WriteTo(w io.Writer) (int64, error) {

	err := binary.Write(w, binary.BigEndian, MessageType)
	if err != nil {
		return 0, err
	}

	var n int64 = 1

	err = binary.Write(w, binary.BigEndian, uint32(len(m)))

	if err != nil {
		return n, err
	}

	n += 4

	o, err := w.Write([]byte(m))

	return n + int64(o), err

}

func (m *Message) ReadFrom(r io.Reader) (int64, error) {

	var n int64 = 1

	var size uint32

	err := binary.Read(r, binary.BigEndian, size)

	if err != nil {
		return n, err
	}

	if size > MaxPayloadsize {
		return n, ErrorMaxPayloadSize
	}

	n += 4

	buf := make([]byte, size)

	o, err := r.Read(buf)

	if err != nil {
		return n, err
	}

	*m = Message(buf)

	return n + int64(o), err
}

func Decode(r io.Reader) (Payload, error) {
	var typ uint8
	var payload Payload
	err := binary.Read(r, binary.BigEndian, &typ)

	if err != nil {
		return nil, err
	}

	switch typ {
	case MessageType:
		payload = new(Message)

	default:
		return nil, errors.New("unknown message type")
	}

	_, err = payload.ReadFrom(r)

	if err != nil {
		return nil, err
	}

	return payload, nil
}
