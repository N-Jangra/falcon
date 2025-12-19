# Falcon Tunnel

Go-based secure FTP tunnel service. Sprint 0 delivers the project skeleton, module setup, and baseline documentation to start building.

## Status
- Sprint 0 complete: project structure, Go module, core dependencies (logrus, bcrypt, yaml), and initial docs.
- No functional tunnel logic yet; server and client entrypoints are placeholders.

## Getting Started
```bash
go mod download
go test ./...
```

Run with example config:
```bash
go run ./cmd/server --config config.example.yaml
go run ./cmd/client --config config.example.yaml
```

The server accepts TCP connections (optionally over TLS), performs an auth handshake, and proxies bytes to the configured FTP server. The client listens on a local FTP port and forwards traffic through the tunnel to the server.

## Project Structure
- `cmd/server` and `cmd/client` - CLI entrypoints
- `internal/auth` - Authentication helpers (bcrypt scaffolding)
- `internal/config` - Config structures, YAML loader, CLI overrides, validation
- `internal/logger` - Structured logging with level/format and optional file output
- `internal/tunnel` - Server and client implementations with proxy logic (TCP/TLS)
- `pkg/protocol` - Protocol message definitions, encoding/decoding, framing
- `docs/` - Requirements and development notes
- `config.example.yaml` - Sample configuration
- `cmd/tlsgen` - Self-signed certificate generation utility

## Documentation
- Requirements: `docs/requirements.md`
- Development setup: `docs/development.md`
- Protocol: `docs/protocol.md`
- TLS: `docs/tls.md`
- Sprint plan: `sprints.md`

## Next Steps
- Implement Sprint 6: connection management (timeouts, pooling, reconnection).
