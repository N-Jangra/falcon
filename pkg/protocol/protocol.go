package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// MessageType represents the type of protocol message exchanged over the tunnel.
type MessageType byte

const (
	MsgAuth MessageType = iota
	MsgAuthResponse
	MsgData
	MsgClose
	MsgHeartbeat
)

// Message is the base frame type for tunnel communication.
// Payload semantics:
// - MsgAuth: plaintext password or token (later)
// - MsgAuthResponse: "ok" or error string
// - MsgData: raw FTP payload
// - MsgClose: optional reason text
// - MsgHeartbeat: empty payload
type Message struct {
	Type    MessageType
	Payload []byte
}

const (
	headerSize   = 5 // 1 byte type + 4 byte payload length
	maxPayload   = 1 << 20
	minFrameSize = headerSize
)

var (
	// ErrFrameTooLarge signals a payload exceeds allowed size.
	ErrFrameTooLarge = errors.New("protocol: frame too large")
	// ErrIncompleteFrame signals a frame that cannot be fully read.
	ErrIncompleteFrame = errors.New("protocol: incomplete frame")
)

// Encode encodes a Message into a length-prefixed frame: [1 byte type][4 byte payload length][payload].
func Encode(msg Message) ([]byte, error) {
	if len(msg.Payload) > maxPayload {
		return nil, ErrFrameTooLarge
	}
	buf := bytes.NewBuffer(make([]byte, 0, headerSize+len(msg.Payload)))
	buf.WriteByte(byte(msg.Type))
	if err := binary.Write(buf, binary.BigEndian, uint32(len(msg.Payload))); err != nil {
		return nil, err
	}
	buf.Write(msg.Payload)
	return buf.Bytes(), nil
}

// Decode reads a single frame from r and returns a Message.
// Caller should wrap r with a deadline on network connections.
func Decode(r io.Reader) (*Message, error) {
	header := make([]byte, headerSize)
	if _, err := io.ReadFull(r, header); err != nil {
		if errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, ErrIncompleteFrame
		}
		return nil, err
	}
	msgType := MessageType(header[0])
	length := binary.BigEndian.Uint32(header[1:])
	if length > maxPayload {
		return nil, ErrFrameTooLarge
	}
	payload := make([]byte, length)
	if length > 0 {
		if _, err := io.ReadFull(r, payload); err != nil {
			if errors.Is(err, io.ErrUnexpectedEOF) {
				return nil, ErrIncompleteFrame
			}
			return nil, err
		}
	}
	return &Message{Type: msgType, Payload: payload}, nil
}

// MustEncode wraps Encode and panics on error; useful for static messages in tests.
func MustEncode(msg Message) []byte {
	b, err := Encode(msg)
	if err != nil {
		panic(fmt.Sprintf("encode message: %v", err))
	}
	return b
}
