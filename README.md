# Go WebSocket Proxy

Web API bridge for CLI tools — exposes system commands via REST, WebSocket, and SSE interfaces.

Designed for scenarios where software has no web interface but needs to be integrated into dashboards and monitoring systems.

## Stack

- **Go** — standard library `net/http`
- **WebSocket** — bidirectional real-time command execution (`nhooyr.io/websocket`)
- **SSE** — streaming command output to browser
- **REST** — synchronous command execution
- **Docker** — multi-stage build

## Quick Start

```bash
# Run directly
go run ./cmd/server

# Or with Docker
docker compose up -d
```

## API Endpoints

### REST

```bash
# Health check
curl http://localhost:8080/api/health

# Execute command
curl -X POST http://localhost:8080/api/exec \
  -H "Content-Type: application/json" \
  -d '{"command": "ping", "args": ["-c", "4", "8.8.8.8"]}'
```

### WebSocket

```
ws://localhost:8080/ws
```

Send JSON messages:
```json
{"action": "exec", "command": "hostname", "args": []}
{"action": "stream", "command": "ping", "args": ["-c", "4", "8.8.8.8"]}
```

### SSE (Server-Sent Events)

```bash
# Stream command output
curl http://localhost:8080/sse/exec?command=ping&args=-c,4,8.8.8.8
```

## Architecture

```
cmd/server/          # Entry point, routing
internal/
├── api/             # REST handlers
├── websocket/       # WebSocket handler (bidirectional)
├── sse/             # SSE handler (server → client streaming)
└── proxy/           # Command executor with allowlist
configs/             # Environment-based configuration
```

## Security

Commands are restricted to an allowlist defined in `internal/proxy/executor.go`. Only explicitly allowed commands can be executed through any interface.

## TODO

- [ ] Configurable command allowlist via config file
- [ ] Authentication (API key or JWT)
- [ ] Rate limiting per connection
- [ ] Metrics endpoint (Prometheus)
- [ ] Graceful shutdown
