package tunnel

import (
	"context"
	"crypto/tls"
	"fmt"
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
	tlsCfg  *tls.Config
	mu      sync.Mutex
	running bool
}

// NewClient constructs a Client with defaults.
func NewClient(cfg config.Config, logger *logrus.Logger, tlsCfg *tls.Config) *Client {
	if logger == nil {
		logger = logrus.New()
	}
	return &Client{
		cfg:    cfg,
		logger: logger,
		tlsCfg: tlsCfg,
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

	tunnelConn, err := c.dialWithRetry(ctx)
	if err != nil {
		return err
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

	return proxyWithIdle(localConn, tunnelConn, c.cfg.Client.IdleTimeout)
}

func (c *Client) dialWithRetry(ctx context.Context) (net.Conn, error) {
	attempts := c.cfg.Client.MaxRetries
	if attempts == 0 {
		attempts = 1
	}
	backoff := c.cfg.Client.BackoffInitial
	if backoff == 0 {
		backoff = 500 * time.Millisecond
	}
	maxBackoff := c.cfg.Client.BackoffMax
	if maxBackoff == 0 {
		maxBackoff = 5 * time.Second
	}

	for i := 0; ; i++ {
		conn, err := c.dialOnce(ctx)
		if err == nil {
			return conn, nil
		}
		if i+1 >= attempts {
			return nil, fmt.Errorf("dial tunnel server: %w", err)
		}
		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}
}

func (c *Client) dialOnce(ctx context.Context) (net.Conn, error) {
	dialTimeout := c.cfg.Client.Timeout
	if dialTimeout == 0 {
		dialTimeout = 30 * time.Second
	}
	dialer := &net.Dialer{Timeout: dialTimeout, KeepAlive: c.cfg.Client.KeepAlive}
	if c.tlsCfg != nil {
		tlsDialer := tls.Dialer{
			NetDialer: dialer,
			Config:    c.tlsCfg,
		}
		return tlsDialer.DialContext(ctx, "tcp", c.cfg.Client.TunnelAddr)
	}
	return dialer.DialContext(ctx, "tcp", c.cfg.Client.TunnelAddr)
}
