# Development Environment

## Prerequisites
- Go 1.24+ (toolchain auto-selects go1.24.11 via `go` if available)
- Git

## Initial Setup
1) Clone the repository  
2) From the project root run:
```bash
go mod download
go test ./...
```

## Running with the sample config
```bash
go run ./cmd/server --config config.example.yaml
go run ./cmd/client --config config.example.yaml
```

## Project Layout
- `cmd/` - CLI entrypoints for server and client
- `internal/` - Application packages (auth, config, logger, tunnel)
- `pkg/` - Shared protocol definitions
- `docs/` - Requirements, design, and operational docs
- `docs/protocol.md` - Message framing and auth handshake reference

## Coding Standards
- `gofmt` before committing (already configured via Go toolchain).
- Prefer structured logging with logrus.
- Keep configuration in YAML; mirror fields with CLI flags when added.

## Tooling Tips
- Run tests with `go test ./...`
- Add new dependencies with `go get <module>@latest` and commit `go.mod`/`go.sum`.
- Use `GOWORK=off` unless a workspace is explicitly added.
