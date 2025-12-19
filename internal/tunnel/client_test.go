package tunnel

import (
	"context"
	"crypto/tls"
	"net"
	"os"
	"path/filepath"
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

	serverLogger := logrus.New()
	serverLogger.SetLevel(logrus.DebugLevel)
	server := NewServer(serverCfg, nil, serverLogger)
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

	client := NewClient(clientCfg, logrus.New(), nil)
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

func TestClientEndToEndEchoTLS(t *testing.T) {
	// Fake FTP echo server
	ftpLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("ftp listen: %v", err)
	}
	defer ftpLn.Close()
	go acceptAndEcho(ftpLn)

	// TLS assets
	cert, key, err := config.GenerateSelfSigned("127.0.0.1", time.Hour)
	if err != nil {
		t.Fatalf("self-signed: %v", err)
	}
	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")
	if err := os.WriteFile(certPath, cert, 0o644); err != nil {
		t.Fatalf("write cert: %v", err)
	}
	if err := os.WriteFile(keyPath, key, 0o600); err != nil {
		t.Fatalf("write key: %v", err)
	}

	// Tunnel server
	serverLn, err := tls.Listen("tcp", "127.0.0.1:0", mustServerTLSConfig(t, certPath, keyPath))
	if err != nil {
		t.Fatalf("server listen tls: %v", err)
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
		TLS: config.TLSConfig{
			Enabled:  true,
			CertFile: certPath,
			KeyFile:  keyPath,
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
		TLS: config.TLSConfig{
			Enabled:    true,
			CAFile:     certPath,
			ServerName: "127.0.0.1",
		},
		Log: config.LogConfig{
			Level:  "error",
			Format: "text",
		},
	}

	clientTLS, err := config.ClientTLSConfig(clientCfg.TLS)
	if err != nil {
		t.Fatalf("client tls config: %v", err)
	}

	clientLogger := logrus.New()
	clientLogger.SetLevel(logrus.DebugLevel)
	client := NewClient(clientCfg, clientLogger, clientTLS)
	clientCtx, clientCancel := context.WithCancel(context.Background())
	defer clientCancel()
	go func() {
		_ = client.Start(clientCtx)
	}()

	time.Sleep(100 * time.Millisecond)

	ftpClient, err := net.Dial("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(localPort)))
	if err != nil {
		t.Fatalf("ftp client dial: %v", err)
	}
	defer ftpClient.Close()

	payload := []byte("hello through tls tunnel")
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

func mustServerTLSConfig(t *testing.T, certPath, keyPath string) *tls.Config {
	t.Helper()
	cfg, err := config.ServerTLSConfig(config.TLSConfig{
		Enabled:  true,
		CertFile: certPath,
		KeyFile:  keyPath,
	})
	if err != nil {
		t.Fatalf("server tls config: %v", err)
	}
	return cfg
}
