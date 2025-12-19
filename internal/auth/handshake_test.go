package auth

import (
	"errors"
	"net"
	"testing"
	"time"
)

func TestHandshakeSuccess(t *testing.T) {
	pass := "secret"
	hash, err := HashPassword(pass)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	serverAuth := New(hash)

	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	go func() {
		_ = HandshakeServer(serverConn, serverAuth, time.Second)
	}()

	if err := HandshakeClient(clientConn, pass, time.Second); err != nil {
		t.Fatalf("client handshake: %v", err)
	}
}

func TestHandshakeFailure(t *testing.T) {
	pass := "secret"
	hash, err := HashPassword(pass)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	serverAuth := New(hash)

	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	go func() {
		_ = HandshakeServer(serverConn, serverAuth, time.Second)
	}()

	if err := HandshakeClient(clientConn, "badpass", time.Second); !errors.Is(err, ErrAuthFailed) {
		t.Fatalf("expected auth failure, got %v", err)
	}
}

func TestGenerateToken(t *testing.T) {
	var calls int
	fakeRand := func(b []byte) (int, error) {
		calls++
		for i := range b {
			b[i] = byte(i + 1)
		}
		return len(b), nil
	}
	token, err := GenerateToken(4, fakeRand)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	if token != "01020304" {
		t.Fatalf("expected hex token, got %s", token)
	}
	if calls != 1 {
		t.Fatalf("expected fakeRand called once")
	}

	if _, err := GenerateToken(0, nil); err == nil {
		t.Fatalf("expected error on zero length token")
	}
}
