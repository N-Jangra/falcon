package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/njangra/falcon-tunnel/pkg/protocol"
)

var (
	// ErrAuthFailed indicates authentication failed.
	ErrAuthFailed = errors.New("authentication failed")
	// ErrInvalidResponse indicates a malformed auth response.
	ErrInvalidResponse = errors.New("invalid auth response")
	// ErrTokenGeneration indicates token generation failure.
	ErrTokenGeneration = errors.New("failed generating session token")
)

// HandshakeServer performs a simple password-based authentication handshake.
// It expects an auth message containing the plaintext password, verifies it,
// and responds with MsgAuthResponse ("ok" or error text).
func HandshakeServer(conn net.Conn, authenticator *Authenticator, deadline time.Duration) error {
	if deadline > 0 {
		_ = conn.SetDeadline(time.Now().Add(deadline))
	}
	defer conn.SetDeadline(time.Time{})

	msg, err := protocol.Decode(conn)
	if err != nil {
		return fmt.Errorf("read auth message: %w", err)
	}
	if msg.Type != protocol.MsgAuth {
		return ErrInvalidResponse
	}
	password := string(msg.Payload)
	var response []byte
	if authenticator.Authenticate(password) {
		response = []byte("ok")
	} else {
		response = []byte("invalid credentials")
	}

	respFrame, err := protocol.Encode(protocol.Message{Type: protocol.MsgAuthResponse, Payload: response})
	if err != nil {
		return err
	}
	if _, err := conn.Write(respFrame); err != nil {
		return fmt.Errorf("write auth response: %w", err)
	}

	if string(response) != "ok" {
		return ErrAuthFailed
	}
	return nil
}

// HandshakeClient sends plaintext password and expects "ok" in response.
func HandshakeClient(conn net.Conn, password string, deadline time.Duration) error {
	if deadline > 0 {
		_ = conn.SetDeadline(time.Now().Add(deadline))
	}
	defer conn.SetDeadline(time.Time{})

	frame, err := protocol.Encode(protocol.Message{Type: protocol.MsgAuth, Payload: []byte(password)})
	if err != nil {
		return err
	}
	if _, err := conn.Write(frame); err != nil {
		return fmt.Errorf("send auth: %w", err)
	}

	resp, err := protocol.Decode(conn)
	if err != nil {
		return fmt.Errorf("read auth response: %w", err)
	}
	if resp.Type != protocol.MsgAuthResponse {
		return ErrInvalidResponse
	}
	if string(resp.Payload) != "ok" {
		return ErrAuthFailed
	}
	return nil
}

// GenerateToken produces a random session token of n bytes (hex encoded).
// The optional randSrc allows deterministic testing; nil uses crypto/rand.Reader.
func GenerateToken(n int, randSrc func([]byte) (int, error)) (string, error) {
	if n <= 0 {
		return "", fmt.Errorf("token length must be >0")
	}
	buf := make([]byte, n)
	read := randSrc
	if read == nil {
		read = rand.Read
	}
	if _, err := read(buf); err != nil {
		return "", fmt.Errorf("%w: %v", ErrTokenGeneration, err)
	}
	return hex.EncodeToString(buf), nil
}
