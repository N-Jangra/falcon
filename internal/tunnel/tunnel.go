package tunnel

import (
	"net"

	"github.com/njangra/falcon-tunnel/internal/auth"
)

// Server is a placeholder tunnel server definition.
type Server struct {
	Listener      net.Listener
	FTPServerAddr string
	Authenticator *auth.Authenticator
}

// Client is a placeholder tunnel client definition.
type Client struct {
	TunnelAddr string
	LocalPort  int
}
