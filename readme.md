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

The server now accepts TCP connections, performs an auth handshake, and proxies bytes to the configured FTP server.

## Project Structure
- `cmd/server` and `cmd/client` - CLI entrypoints
- `internal/auth` - Authentication helpers (bcrypt scaffolding)
- `internal/config` - Config structures, YAML loader, CLI overrides, validation
- `internal/logger` - Structured logging with level/format and optional file output
- `internal/tunnel` - Server/client placeholders
- `pkg/protocol` - Protocol message definitions, encoding/decoding, framing
- `docs/` - Requirements and development notes
- `config.example.yaml` - Sample configuration

## Documentation
- Requirements: `docs/requirements.md`
- Development setup: `docs/development.md`
- Protocol: `docs/protocol.md`
- Sprint plan: `sprints.md`

## Next Steps
- Implement Sprint 4: basic tunnel client and local FTP listener for end-to-end flow.
