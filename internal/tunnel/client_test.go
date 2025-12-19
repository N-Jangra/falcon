package tunnel

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/njangra/falcon-tunnel/internal/auth"
	"github.com/njangra/falcon-tunnel/internal/config"
	"github.com/sirupsen/logrus"
)

func TestClientEndToEndEcho(t *testing.T) {
	// Fake FTP echo server
	ftpLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("ftp listen: %v", err)
	}
	defer ftpLn.Close()
	go acceptAndEcho(ftpLn)

	// Tunnel server
	serverLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("server listen: %v", err)
	}
	defer serverLn.Close()

	hash, err := auth.HashPassword("secret")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	serverCfg := config.Config{
		Server: config.ServerConfig{
			ListenAddr:     serverLn.Addr().String(),
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

	server := NewServer(serverCfg, nil, logrus.New())
	serverCtx, serverCancel := context.WithCancel(context.Background())
	defer serverCancel()
	go func() {
		_ = server.Serve(serverCtx, serverLn)
	}()

	// Pick local port for client listener.
	localPort := pickFreePort(t)

	clientCfg := config.Config{
		Client: config.ClientConfig{
			TunnelAddr:   serverLn.Addr().String(),
			LocalFTPPort: localPort,
			Timeout:      2 * time.Second,
			Password:     "secret",
		},
		Auth: config.AuthConfig{
			Enabled: true,
		},
		Log: config.LogConfig{
			Level:  "error",
			Format: "text",
		},
	}

	client := NewClient(clientCfg, logrus.New())
	clientCtx, clientCancel := context.WithCancel(context.Background())
	defer clientCancel()
	go func() {
		_ = client.Start(clientCtx)
	}()

	// Wait briefly for client listener to start.
	time.Sleep(100 * time.Millisecond)

	// Simulate FTP client connecting to local listener.
	ftpClient, err := net.Dial("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(localPort)))
	if err != nil {
		t.Fatalf("ftp client dial: %v", err)
	}
	defer ftpClient.Close()

	payload := []byte("hello through tunnel")
	if _, err := ftpClient.Write(payload); err != nil {
		t.Fatalf("write payload: %v", err)
	}

	buf := make([]byte, len(payload))
	if _, err := ftpClient.Read(buf); err != nil {
		t.Fatalf("read response: %v", err)
	}
	if string(buf) != string(payload) {
		t.Fatalf("expected echo %q got %q", payload, buf)
	}
}

func pickFreePort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("pick free port: %v", err)
	}
	defer ln.Close()
	return ln.Addr().(*net.TCPAddr).Port
}
