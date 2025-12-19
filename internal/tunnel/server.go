package tunnel

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"

	"github.com/njangra/falcon-tunnel/internal/auth"
	"github.com/njangra/falcon-tunnel/internal/config"
	"github.com/sirupsen/logrus"
)

// Server is a minimal TCP tunnel server that authenticates clients and proxies bytes to an FTP server.
type Server struct {
	cfg           config.Config
	authenticator *auth.Authenticator
	logger        *logrus.Logger

	conns   map[uint64]net.Conn
	connMu  sync.Mutex
	nextID  uint64
	maxConn int
	wg      sync.WaitGroup

	pool *connPool
}

// NewServer constructs a Server.
func NewServer(cfg config.Config, authenticator *auth.Authenticator, logger *logrus.Logger) *Server {
	if authenticator == nil {
		authenticator = auth.New(cfg.Auth.PasswordHash)
	}
	if logger == nil {
		logger = logrus.New()
	}
	pool := newConnPool(cfg.Server.FTPServerAddr, cfg.Server.Timeout, cfg.Server.IdleTimeout, cfg.Server.PoolSize)
	return &Server{
		cfg:           cfg,
		authenticator: authenticator,
		logger:        logger,
		conns:         make(map[uint64]net.Conn),
		maxConn:       cfg.Server.MaxConnections,
		pool:          pool,
	}
}

// Serve begins accepting connections on the provided listener until ctx is cancelled.
func (s *Server) Serve(ctx context.Context, ln net.Listener) error {
	defer s.shutdown()
	defer s.pool.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) || ctx.Err() != nil {
				return nil
			}
			s.logger.WithError(err).Warn("accept failed")
			continue
		}

		if !s.registerConn(conn) {
			s.logger.WithField("remote", conn.RemoteAddr().String()).Warn("max connections reached, closing connection")
			_ = conn.Close()
			continue
		}

		s.wg.Add(1)
		go func(c net.Conn) {
			defer s.wg.Done()
			defer s.unregisterConn(c)
			if err := s.handleConn(ctx, c); err != nil {
				s.logger.WithError(err).Debug("connection closed with error")
			}
		}(conn)
	}
}

func (s *Server) handleConn(ctx context.Context, tunnelConn net.Conn) error {
	if s.cfg.Auth.Enabled {
		if err := auth.HandshakeServer(tunnelConn, s.authenticator, s.cfg.Server.Timeout); err != nil {
			return fmt.Errorf("auth handshake: %w", err)
		}
	}

	s.logger.WithField("ftp", s.cfg.Server.FTPServerAddr).Debug("acquiring ftp connection")
	acquireCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Server.Timeout)
	defer cancel()
	ftpConn, err := s.pool.Acquire(acquireCtx)
	if err != nil {
		return fmt.Errorf("dial ftp server: %w", err)
	}
	s.logger.WithField("ftp", s.cfg.Server.FTPServerAddr).Debug("ftp connection acquired")
	if ftpConn == nil {
		return fmt.Errorf("dial ftp server: nil connection")
	}

	s.logger.WithFields(logrus.Fields{
		"remote": tunnelConn.RemoteAddr().String(),
		"ftp":    s.cfg.Server.FTPServerAddr,
	}).Info("proxy connection established")

	copyErr := proxyWithIdle(tunnelConn, ftpConn, s.cfg.Server.IdleTimeout)
	s.pool.Release(ftpConn, copyErr == nil)
	return copyErr
}

func proxy(a, b net.Conn) error {
	errs := make(chan error, 2)
	// Copy in both directions and close the opposite side when done.
	go func() {
		_, err := io.Copy(a, b)
		_ = a.Close()
		errs <- err
	}()
	go func() {
		_, err := io.Copy(b, a)
		_ = b.Close()
		errs <- err
	}()

	// Return the first non-nil error (if any).
	var firstErr error
	for i := 0; i < 2; i++ {
		if err := <-errs; err != nil && firstErr == nil && !errors.Is(err, net.ErrClosed) {
			firstErr = err
		}
	}
	return firstErr
}

func (s *Server) registerConn(c net.Conn) bool {
	s.connMu.Lock()
	defer s.connMu.Unlock()
	if s.maxConn > 0 && len(s.conns) >= s.maxConn {
		return false
	}
	id := atomic.AddUint64(&s.nextID, 1)
	s.conns[id] = c
	return true
}

func (s *Server) unregisterConn(c net.Conn) {
	s.connMu.Lock()
	defer s.connMu.Unlock()
	for id, conn := range s.conns {
		if conn == c {
			delete(s.conns, id)
			return
		}
	}
}

func (s *Server) shutdown() {
	s.connMu.Lock()
	for _, c := range s.conns {
		_ = c.Close()
	}
	s.conns = make(map[uint64]net.Conn)
	s.connMu.Unlock()
	s.wg.Wait()
}
