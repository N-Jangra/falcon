# FTP Tunnel App - Workflow Guide

I'll help you create a secure FTP tunnel application in Go. Let me break this down into phases with important features.

## Important Features (Must Have)

1. **Bidirectional Tunnel** - Allow connections from both endpoints
2. **Authentication** - Password protection for tunnel access
3. **Encryption** - TLS/SSL for secure communication
4. **Connection Pooling** - Reuse connections efficiently
5. **Error Handling & Reconnection** - Automatic retry logic
6. **Logging** - Track connections and errors
7. **Configuration** - Easy setup via config file/flags

## Good to Have Features

1. **Multiple simultaneous connections** - Handle concurrent FTP sessions
2. **Bandwidth monitoring** - Track data transfer rates
3. **Connection timeout** - Prevent hanging connections
4. **Whitelist/Blacklist** - IP filtering
5. **Health checks** - Monitor tunnel status
6. **Compression** - Reduce bandwidth usage
7. **Web dashboard** - Monitor tunnel status via HTTP
8. **Session management** - Track active sessions

## Go Implementation Workflow

### **Phase 1: Project Structure & Setup**

```
ftp-tunnel/
├── cmd/
│   ├── server/
│   │   └── main.go          # Server entry point
│   └── client/
│       └── main.go          # Client entry point
├── internal/
│   ├── auth/
│   │   └── auth.go          # Authentication logic
│   ├── tunnel/
│   │   ├── server.go        # Tunnel server
│   │   ├── client.go        # Tunnel client
│   │   └── connection.go    # Connection handling
│   ├── config/
│   │   └── config.go        # Configuration management
│   └── logger/
│       └── logger.go        # Logging utilities
├── pkg/
│   └── protocol/
│       └── protocol.go      # Custom protocol definitions
├── config.yaml              # Configuration file
└── go.mod
```

**Tasks:**
- Initialize Go module: `go mod init ftp-tunnel`
- Install dependencies:
  - `go get golang.org/x/crypto/bcrypt` (password hashing)
  - `go get gopkg.in/yaml.v3` (config parsing)
  - `go get github.com/sirupsen/logrus` (logging)

### **Phase 2: Configuration Management**

Create a flexible configuration system:

```go
// internal/config/config.go
type Config struct {
    Server ServerConfig
    Client ClientConfig
    Auth   AuthConfig
    TLS    TLSConfig
}

type ServerConfig struct {
    ListenAddr    string
    FTPServerAddr string
    MaxConnections int
    Timeout       time.Duration
}

type ClientConfig struct {
    TunnelAddr string
    LocalFTPPort int
}

type AuthConfig struct {
    Enabled  bool
    Password string // bcrypt hashed
}
```

**Tasks:**
- Parse YAML/JSON config files
- Support command-line flags override
- Validate configuration values

### **Phase 3: Authentication Module**

Implement secure authentication:

```go
// internal/auth/auth.go
type Authenticator struct {
    passwordHash string
}

func (a *Authenticator) Authenticate(password string) bool
func (a *Authenticator) HashPassword(password string) string
```

**Tasks:**
- Use bcrypt for password hashing
- Implement challenge-response authentication
- Add token-based session management (optional)

### **Phase 4: Protocol Definition**

Define custom protocol for tunnel communication:

```go
// pkg/protocol/protocol.go
type MessageType byte

const (
    MsgAuth MessageType = iota
    MsgAuthResponse
    MsgData
    MsgClose
    MsgHeartbeat
)

type Message struct {
    Type    MessageType
    Payload []byte
}

func EncodeMessage(msg Message) []byte
func DecodeMessage(data []byte) (Message, error)
```

**Tasks:**
- Define message types and formats
- Implement serialization/deserialization
- Add message framing (length prefix)

### **Phase 5: Tunnel Server Implementation**

Create the server that accepts tunnel connections:

```go
// internal/tunnel/server.go
type TunnelServer struct {
    listener      net.Listener
    ftpAddr       string
    authenticator *auth.Authenticator
    connections   sync.Map
}

func (s *TunnelServer) Start() error
func (s *TunnelServer) handleConnection(conn net.Conn)
func (s *TunnelServer) proxyToFTP(tunnelConn, ftpConn net.Conn)
```

**Workflow:**
1. Listen on tunnel port
2. Accept incoming connections
3. Perform authentication handshake
4. Connect to local FTP server
5. Bidirectional proxy data between tunnel and FTP
6. Handle connection cleanup

**Tasks:**
- Implement TLS listener
- Add connection pooling
- Handle multiple concurrent connections
- Implement graceful shutdown

### **Phase 6: Tunnel Client Implementation**

Create the client that connects to the tunnel:

```go
// internal/tunnel/client.go
type TunnelClient struct {
    serverAddr string
    localPort  int
    password   string
}

func (c *TunnelClient) Start() error
func (c *TunnelClient) connectToServer() (net.Conn, error)
func (c *TunnelClient) authenticate(conn net.Conn) error
func (c *TunnelClient) handleLocalConnection(local, tunnel net.Conn)
```

**Workflow:**
1. Listen on local FTP port
2. Accept local FTP client connections
3. Establish tunnel connection to server
4. Authenticate with server
5. Proxy data bidirectionally
6. Reconnect on tunnel failure

**Tasks:**
- Implement automatic reconnection
- Add connection retry logic with backoff
- Handle authentication failures
- Implement keep-alive mechanism

### **Phase 7: Bidirectional Data Proxying**

Implement efficient data transfer:

```go
// internal/tunnel/connection.go
type Connection struct {
    client net.Conn
    server net.Conn
    closeOnce sync.Once
}

func (c *Connection) Proxy() error {
    go c.copyData(c.client, c.server)
    go c.copyData(c.server, c.client)
}

func (c *Connection) copyData(dst, src net.Conn) error
```

**Tasks:**
- Use `io.Copy` for efficient transfer
- Handle partial reads/writes
- Implement proper error handling
- Add bandwidth tracking
- Implement connection timeout

### **Phase 8: TLS/Encryption**

Add transport security:

```go
// internal/config/tls.go
func LoadTLSConfig(certFile, keyFile string) (*tls.Config, error)
func GenerateSelfSignedCert() (tls.Certificate, error)
```

**Tasks:**
- Generate or load TLS certificates
- Configure TLS for server and client
- Support self-signed certificates option
- Implement certificate pinning (optional)

### **Phase 9: Logging & Monitoring**

Implement comprehensive logging:

```go
// internal/logger/logger.go
type Logger struct {
    *logrus.Logger
}

func (l *Logger) LogConnection(addr string)
func (l *Logger) LogError(err error)
func (l *Logger) LogTransfer(bytes int64, duration time.Duration)
```

**Tasks:**
- Log all connection events
- Track data transfer statistics
- Add structured logging
- Implement log rotation
- Add metrics collection (optional)

### **Phase 10: Advanced Features**

Implement nice-to-have features:

**Connection Management:**
```go
type ConnectionManager struct {
    active    map[string]*Connection
    maxConns  int
    bandwidth *BandwidthMonitor
}
```

**Health Checks:**
```go
func (s *TunnelServer) HealthCheck() error
func (c *TunnelClient) IsConnected() bool
```

**Web Dashboard:**
```go
func (s *TunnelServer) StartWebDashboard(port int)
// Serve status, active connections, bandwidth stats
```

### **Phase 11: Testing**

Create comprehensive tests:

```go
// internal/tunnel/server_test.go
func TestAuthentication(t *testing.T)
func TestDataProxy(t *testing.T)
func TestReconnection(t *testing.T)
```

**Tasks:**
- Unit tests for each component
- Integration tests for end-to-end flow
- Load testing for concurrent connections
- Test failure scenarios

### **Phase 12: Documentation & Deployment**

Final touches:

**Tasks:**
- Create README with setup instructions
- Add example configuration files
- Build binaries for different platforms
- Create Docker containers (optional)
- Add systemd service files

## Example Usage

**Server:**
```bash
./ftp-tunnel-server --config config.yaml
# or
./ftp-tunnel-server --listen :8080 --ftp localhost:21 --password secret
```

**Client:**
```bash
./ftp-tunnel-client --server tunnel.example.com:8080 --local-port 2121 --password secret
```

**FTP Client Connection:**
```bash
ftp localhost 2121
# Traffic goes through encrypted tunnel to remote FTP server
```

## Key Go Packages to Use

- `net` - TCP connections
- `crypto/tls` - TLS encryption
- `io` - Data copying
- `sync` - Concurrency primitives
- `context` - Cancellation and timeouts
- `golang.org/x/crypto/bcrypt` - Password hashing
- `gopkg.in/yaml.v3` - Config parsing

Would you like me to create a starter implementation with any specific phase or feature?