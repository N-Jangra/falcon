package tunnel

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/njangra/falcon-tunnel/internal/auth"
	"github.com/njangra/falcon-tunnel/internal/config"
	"github.com/sirupsen/logrus"
)

func TestServerProxiesData(t *testing.T) {
	// Start fake FTP echo server.
	ftpLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("ftp listen: %v", err)
	}
	defer ftpLn.Close()
	go acceptAndEcho(ftpLn)

	hash, err := auth.HashPassword("secret")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	cfg := config.Config{
		Server: config.ServerConfig{
			ListenAddr:     "127.0.0.1:0",
			FTPServerAddr:  ftpLn.Addr().String(),
			MaxConnections: 10,
			Timeout:        2 * time.Second,
		},
		Auth: config.AuthConfig{
			Enabled:      true,
			PasswordHash: hash,
		},
		Log: config.LogConfig{
			Level:  "error",
			Format: "text",
		},
	}

	ln, err := net.Listen("tcp", cfg.Server.ListenAddr)
	if err != nil {
		t.Fatalf("server listen: %v", err)
	}
	defer ln.Close()

	srv := NewServer(cfg, nil, logrus.New())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		_ = srv.Serve(ctx, ln)
	}()

	// Client connects to tunnel server.
	clientConn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("client dial: %v", err)
	}
	defer clientConn.Close()

	if err := auth.HandshakeClient(clientConn, "secret", cfg.Server.Timeout); err != nil {
		t.Fatalf("handshake client: %v", err)
	}

	payload := []byte("ping over tunnel")
	if _, err := clientConn.Write(payload); err != nil {
		t.Fatalf("write payload: %v", err)
	}

	buf := make([]byte, len(payload))
	if _, err := clientConn.Read(buf); err != nil {
		t.Fatalf("read response: %v", err)
	}
	if string(buf) != string(payload) {
		t.Fatalf("expected echo %q got %q", payload, buf)
	}

	cancel()
}

func acceptAndEcho(ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		go func(c net.Conn) {
			defer c.Close()
			buf := make([]byte, 1024)
			for {
				n, err := c.Read(buf)
				if err != nil {
					return
				}
				_, _ = c.Write(buf[:n])
			}
		}(conn)
	}
}
