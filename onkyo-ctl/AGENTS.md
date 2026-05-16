## Architecture

Two binaries sharing `internal/pkg/eiscp` and `internal/config`:
- **API** (`cmd/api/`) — production HTTP server, holds the single persistent EISCP connection to the Onkyo receiver. LAN devices talk to it via REST.
- **CLI** (`cmd/cli/`) — testing/debug tool. Direct EISCP connection to Onkyo.

The Onkyo receiver accepts **at most one TCP connection at a time**. The API exists to own that connection so multiple LAN clients can interact asynchronously.

**Endpoints:**
- Onkyo receiver: `10.204.0.163:60128`
- API server (LAN): `10.205.0.5:8001` (Docker internal: `:8080`)

## File bloom filter

### internal/config/config.go
- **Could contain:** profile definitions, YAML config loading, host/port defaults
- **Definitely NOT:** EISCP protocol, HTTP handling, CLI flags

### internal/pkg/eiscp/client.go
- **Could contain:** ISCP command implementations, TCP connection management, response parsing
- **Definitely NOT:** HTTP routing, config loading, CLI framework

### internal/pkg/eiscp/packet.go
- **Could contain:** eISCP binary wire format, serialization, header structure
- **Definitely NOT:** high-level commands, error types

### cmd/api/main.go
- **Could contain:** chi router, HTTP handlers for power/volume/subwoofer/input/profile, health endpoint
- **Definitely NOT:** EISCP packet framing, config defaults

### cmd/cli/main.go
- **Could contain:** urfave/cli command tree, flag/env override for host/port
- **Definitely NOT:** HTTP, REST endpoints

### cmd/cli/chat.go
- **Could contain:** interactive readline loop, raw ISCP command entry
- **Definitely NOT:** structured CLI subcommands, config

## Testing procedure

1. Verify the API is down: `GET http://10.205.0.5:8001/health` (2s timeout)
2. **If reachable** (200 `{"status":"ok"}`) — STOP. API is running and holding the Onkyo connection. Warn user.
3. **If unreachable** (timeout, connection refused) — API is down. Safe to use the CLI for direct testing.
4. **Ask user** before issuing any command that modifies Onkyo state (power, volume, input, subwoofer, brightness).
5. For read-only testing without modification, connect and review query outputs.

## Rules

- **NEVER** edit `go.mod` or `go.sum` directly. Ask the user to run `go get <pkg>` or `go mod tidy`.
