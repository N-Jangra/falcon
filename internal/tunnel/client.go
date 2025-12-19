package tunnel

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/njangra/falcon-tunnel/internal/auth"
	"github.com/njangra/falcon-tunnel/internal/config"
	"github.com/sirupsen/logrus"
)

// Client dials a tunnel server and forwards local FTP connections through it.
type Client struct {
	cfg     config.Config
	logger  *logrus.Logger
	mu      sync.Mutex
	running bool
}

// NewClient constructs a Client with defaults.
func NewClient(cfg config.Config, logger *logrus.Logger) *Client {
	if logger == nil {
		logger = logrus.New()
	}
	return &Client{
		cfg:    cfg,
		logger: logger,
	}
}

// Start begins listening on the configured local FTP port and forwarding connections.
func (c *Client) Start(ctx context.Context) error {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return fmt.Errorf("client already running")
	}
	c.running = true
	c.mu.Unlock()

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", c.cfg.Client.LocalFTPPort))
	if err != nil {
		return fmt.Errorf("listen local ftp: %w", err)
	}
	defer ln.Close()
	c.logger.Infof("client listening on :%d", c.cfg.Client.LocalFTPPort)

	go func() {
		<-ctx.Done()
		_ = ln.Close()
	}()

	var wg sync.WaitGroup
	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				wg.Wait()
				return nil
			default:
				c.logger.WithError(err).Warn("accept failed")
				continue
			}
		}

		wg.Add(1)
		go func(localConn net.Conn) {
			defer wg.Done()
			if err := c.handleLocalConn(ctx, localConn); err != nil {
				c.logger.WithError(err).Debug("local connection closed with error")
			}
		}(conn)
	}
}

func (c *Client) handleLocalConn(ctx context.Context, localConn net.Conn) error {
	defer localConn.Close()

	dialTimeout := c.cfg.Client.Timeout
	if dialTimeout == 0 {
		dialTimeout = 30 * time.Second
	}
	tunnelConn, err := net.DialTimeout("tcp", c.cfg.Client.TunnelAddr, dialTimeout)
	if err != nil {
		return fmt.Errorf("dial tunnel server: %w", err)
	}
	defer tunnelConn.Close()

	if c.cfg.Auth.Enabled {
		if err := auth.HandshakeClient(tunnelConn, c.cfg.Client.Password, c.cfg.Client.Timeout); err != nil {
			return fmt.Errorf("auth handshake: %w", err)
		}
	}

	c.logger.WithFields(logrus.Fields{
		"local":  localConn.RemoteAddr().String(),
		"server": c.cfg.Client.TunnelAddr,
	}).Info("proxy connection established (client)")

	return proxyWithDeadline(localConn, tunnelConn, c.cfg.Client.Timeout)
}

func proxyWithDeadline(a, b net.Conn, timeout time.Duration) error {
	errs := make(chan error, 2)
	copyFunc := func(dst, src net.Conn) {
		if timeout > 0 {
			_ = src.SetDeadline(time.Now().Add(timeout))
			_ = dst.SetDeadline(time.Now().Add(timeout))
		}
		_, err := io.Copy(dst, src)
		_ = dst.Close()
		errs <- err
	}
	go copyFunc(a, b)
	go copyFunc(b, a)

	var firstErr error
	for i := 0; i < 2; i++ {
		if err := <-errs; err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
