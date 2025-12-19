# TLS Setup

## Generating Certificates
- Use the built-in tool to create a self-signed certificate:
```bash
go run ./cmd/tlsgen --host "localhost,127.0.0.1" --cert tls.crt --key tls.key --days 365
```
- For production, use certificates issued by a trusted CA.

## Server Configuration
```yaml
tls:
  enabled: true
  cert_file: "tls.crt"
  key_file: "tls.key"
```
Start the server:
```bash
go run ./cmd/server --config config.example.yaml
```

## Client Configuration
```yaml
tls:
  enabled: true
  ca_file: "tls.crt"        # trust the self-signed cert
  server_name: "localhost"  # must match the certificate CN/SAN
  cert_fingerprint: ""      # optional SHA-256 hex pin
  insecure_skip_verify: false
```
Run the client:
```bash
go run ./cmd/client --config config.example.yaml
```

## Certificate Pinning (Optional)
- Set `tls.cert_fingerprint` to the SHA-256 fingerprint of the server cert (hex, no colons).
- When set, the client will pin the certificate and bypass hostname validation.

## Notes
- Minimum TLS version: 1.2.
- When TLS is enabled the server requires `cert_file` and `key_file`.
- The client can rely on system roots, a provided `ca_file`, pinning, or `insecure_skip_verify` (not recommended).
