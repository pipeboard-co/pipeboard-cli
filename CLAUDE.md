# CLAUDE.md

## Project Overview

pipeboard-cli is a Go CLI for managing ads across Meta, Google, and TikTok via Pipeboard's MCP servers. All operations go through Pipeboard servers — the CLI is a thin client.

**Stack:** Go 1.23 + Cobra CLI framework + goreleaser

## Common Commands

```bash
make build     # Build binary to bin/pipeboard
make install   # Install to $GOPATH/bin
make test      # Run tests
make lint      # Run go vet
```

## Architecture

```
main.go              # Entry point
cmd/                 # Cobra commands
  root.go            # Root command, auth helper, global flags
  version.go         # Version command
  login.go           # Browser OAuth flow (TODO)
  config.go          # Config file management (~/.pipeboard/config.json)
  debug.go           # Hidden debug command for raw API queries
internal/
  client/mcp.go      # MCP JSON-RPC client over HTTP
```

### MCP Protocol

The CLI speaks JSON-RPC 2.0 over HTTP POST to `https://mcp.pipeboard.co/<server-path>`:
- `initialize` — handshake
- `tools/list` — discover available tools
- `tools/call` — invoke a tool
- Auth: `Authorization: Bearer <token>` header

### Auth

Token resolution order:
1. `PIPEBOARD_API_TOKEN` env var
2. `~/.pipeboard/config.json` token field

### Build-time variables

Version and commit are injected via ldflags (see Makefile and .goreleaser.yml):
- `cmd.Version`
- `cmd.Commit`
