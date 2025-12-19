package protocol

import (
	"bytes"
	"errors"
	"testing"
)

func TestEncodeDecodeRoundTrip(t *testing.T) {
	msg := Message{Type: MsgData, Payload: []byte("hello")}
	frame, err := Encode(msg)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	decoded, err := Decode(bytes.NewReader(frame))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if decoded.Type != msg.Type {
		t.Fatalf("expected type %v got %v", msg.Type, decoded.Type)
	}
	if string(decoded.Payload) != "hello" {
		t.Fatalf("payload mismatch")
	}
}

func TestDecodeIncompleteFrame(t *testing.T) {
	frame := []byte{byte(MsgData), 0, 0, 0, 5, 'h', 'i'}
	_, err := Decode(bytes.NewReader(frame))
	if !errors.Is(err, ErrIncompleteFrame) {
		t.Fatalf("expected incomplete frame error, got %v", err)
	}
}

func TestFrameTooLarge(t *testing.T) {
	tooBig := make([]byte, maxPayload+1)
	_, err := Encode(Message{Type: MsgData, Payload: tooBig})
	if !errors.Is(err, ErrFrameTooLarge) {
		t.Fatalf("expected frame too large error")
	}

	header := []byte{byte(MsgData), 0x10, 0x00, 0x00, 0x01} // large length
	_, err = Decode(bytes.NewReader(header))
	if !errors.Is(err, ErrFrameTooLarge) {
		t.Fatalf("expected frame too large on decode")
	}
}
