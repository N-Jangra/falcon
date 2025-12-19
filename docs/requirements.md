# Falcon Tunnel - Requirements

## Overview
Falcon Tunnel is a Go-based secure tunneling service that forwards FTP traffic between a client and server over a managed channel. The system should provide authentication, optional encryption, and operational tooling to run reliably in production or on a developer laptop.

## Goals
- Provide a tunnel that accepts FTP client connections locally and proxies them to a remote FTP server through a controlled TCP channel.
- Offer authentication to prevent unauthorized tunnel use.
- Support TLS so credentials and data can be encrypted in transit.
- Deliver usable defaults with configuration via file and flags.
- Ship minimal operational tooling: logging, metrics hooks, and graceful shutdown.

## Non-Goals
- Implement a full FTP server; the tunnel assumes an existing FTP server endpoint.
- Provide a multi-tenant control plane or web UI beyond the basic dashboard planned in later sprints.
- Persist long-term telemetry or analytics; metrics storage is in-memory only.

## Users & Scenarios
- **Ops engineer**: exposes an internal FTP server to remote users through the tunnel while enforcing authentication and TLS.
- **Developer**: runs the tunnel locally against a staging FTP server to test data flows.
- **Automation**: scheduled jobs transfer files through the tunnel and need predictable retries and logs.

## Functional Requirements
- Tunnel server listens on a configurable TCP address, performs authentication, and proxies FTP data to a target FTP server.
- Tunnel client listens on a local port, establishes the tunnel connection to the server, and forwards local FTP traffic.
- Authentication uses bcrypt-hashed passwords with a simple handshake protocol.
- Configuration can be loaded from YAML and overridden by command-line flags.
- Logging outputs structured logs to console; file output is configurable.
- Graceful shutdown waits for active connections to finish or times out.
- Optional TLS for client/server with self-signed certificate support.
- Basic health checks and metrics endpoints (planned for later sprints).

## Non-Functional Requirements
- **Performance**: Handle at least 50 concurrent FTP connections with acceptable throughput for typical file transfers (tuned in later sprints).
- **Reliability**: Auto-reconnect on transient failures; timeouts protect against hung sessions.
- **Security**: Protect credentials in transit (TLS) and at rest (hashed passwords); support IP filtering and rate limits in later sprints.
- **Observability**: Structured logging; hooks for metrics; optional dashboard in later sprints.
- **Operability**: Configurable via file/flags; usable defaults; clear error messages.

## Interfaces
- **Server CLI**: `./bin/falcon-tunnel-server --config config.yaml --listen :8080 --ftp localhost:21 --password $PASS --tls-cert cert.pem --tls-key key.pem`
- **Client CLI**: `./bin/falcon-tunnel-client --config config.yaml --server localhost:8080 --local-port 2121 --password $PASS --insecure`
- **Config file**: YAML with sections for server, client, auth, tls, logging, and limits.

## Operational Expectations
- Build and run with Go 1.22+.
- Dependency management via Go modules; reproducible builds.
- Provide sample configs and run scripts for local development.

## Risks & Constraints
- FTP active/passive mode handling can be brittle; must be validated during protocol work.
- TLS certificate generation and validation can block connectivity if misconfigured.
- Concurrency bugs and race conditions require careful synchronization and testing.
