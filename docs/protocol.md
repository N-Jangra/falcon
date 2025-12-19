# Falcon Tunnel Protocol

## Framing
- Each frame: 1-byte message type + 4-byte big-endian payload length + payload.
- Max payload size: 1 MiB; larger frames are rejected.
- Heartbeat frames may carry an empty payload.

## Message Types
- `0` - `MsgAuth`: plaintext password or token from client to server.
- `1` - `MsgAuthResponse`: server reply; `"ok"` on success, otherwise error text.
- `2` - `MsgData`: raw FTP data proxied between client and server.
- `3` - `MsgClose`: optional human-readable reason for closing.
- `4` - `MsgHeartbeat`: keep-alive ping/pong (payload optional).

## Authentication Handshake
1. Client sends `MsgAuth` with password.
2. Server validates using bcrypt hash from config.
3. Server replies with `MsgAuthResponse` containing `"ok"` or an error string.
4. On non-`"ok"` responses both sides should close the connection.

## Session Tokens (optional)
- `internal/auth.GenerateToken(n)` creates a random hex-encoded token for future session tracking.
- Tokens can be attached to auth responses or headers in later sprints without changing framing.

## Error Handling Guidelines
- Treat `ErrFrameTooLarge` and `ErrIncompleteFrame` as fatal for the connection.
- Apply read/write deadlines around `Decode`/`Encode` when used over the network.

## Compatibility Notes
- Protocol is binary and versionless for now; changes should be additive or gated by a capability message if added later.
