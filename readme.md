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

## Project Structure
- `cmd/server` and `cmd/client` - CLI entrypoints
- `internal/auth` - Authentication helpers (bcrypt scaffolding)
- `internal/config` - Config structures and YAML loader
- `internal/logger` - Structured logging with logrus
- `internal/tunnel` - Server/client placeholders
- `pkg/protocol` - Protocol message definitions
- `docs/` - Requirements and development notes

## Documentation
- Requirements: `docs/requirements.md`
- Development setup: `docs/development.md`
- Sprint plan: `sprints.md`

## Next Steps
- Implement Sprint 1: configuration loading/validation and logging foundations.
