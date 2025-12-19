package tunnel

import (
	"context"
	"net"
	"time"
)

// connPool is a lightweight semaphore-controlled dialer enforcing max concurrent FTP connections.
// Connections are not reused; this keeps behavior predictable while enforcing resource limits.
type connPool struct {
	target    string
	timeout   time.Duration
	keepAlive time.Duration
	sem       chan struct{}
}

func newConnPool(target string, timeout, keepAlive time.Duration, maxSize int) *connPool {
	if maxSize <= 0 {
		maxSize = 1
	}
	return &connPool{
		target:    target,
		timeout:   timeout,
		keepAlive: keepAlive,
		sem:       make(chan struct{}, maxSize),
	}
}

func (p *connPool) Acquire(ctx context.Context) (net.Conn, error) {
	select {
	case p.sem <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	dialer := &net.Dialer{Timeout: p.timeout, KeepAlive: p.keepAlive}
	conn, err := dialer.DialContext(ctx, "tcp", p.target)
	if err != nil {
		p.releaseToken()
		return nil, err
	}
	return conn, nil
}

func (p *connPool) Release(conn net.Conn, keep bool) {
	if conn != nil {
		if !keep {
			_ = conn.Close()
		} else {
			_ = conn.Close() // close on release; reuse not implemented
		}
	}
	p.releaseToken()
}

func (p *connPool) releaseToken() {
	select {
	case <-p.sem:
	default:
	}
}

func (p *connPool) Close() {
	for {
		select {
		case <-p.sem:
		default:
			return
		}
	}
}
