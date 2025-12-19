package tunnel

import (
	"errors"
	"io"
	"net"
	"time"
)

// proxyWithIdle copies data between two conns and enforces idle timeout via periodic deadline refresh.
func proxyWithIdle(a, b net.Conn, idle time.Duration) error {
	errs := make(chan error, 2)
	stop := make(chan struct{})

	go refreshDeadline(a, idle, stop)
	go refreshDeadline(b, idle, stop)

	go func() {
		_, err := io.Copy(a, b)
		errs <- err
	}()
	go func() {
		_, err := io.Copy(b, a)
		errs <- err
	}()

	var firstErr error
	for i := 0; i < 2; i++ {
		if err := <-errs; err != nil && firstErr == nil && !errors.Is(err, net.ErrClosed) {
			firstErr = err
		}
	}
	close(stop)
	_ = a.Close()
	_ = b.Close()
	return firstErr
}

func refreshDeadline(c net.Conn, idle time.Duration, stop <-chan struct{}) {
	if idle <= 0 {
		return
	}
	t := time.NewTicker(idle / 2)
	defer t.Stop()
	_ = c.SetDeadline(time.Now().Add(idle))
	for {
		select {
		case <-stop:
			return
		case <-t.C:
			_ = c.SetDeadline(time.Now().Add(idle))
		}
	}
}
