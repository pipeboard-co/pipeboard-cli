# pipeboard-cli

Command-line tool for managing Meta Ads, Google Ads, and TikTok Ads from your terminal. Create campaigns, pull reports, and manage ads across platforms with a single CLI.

All operations go through [Pipeboard](https://pipeboard.co) servers — no direct API calls to ad platforms, no credentials stored locally.

## Installation

### Homebrew (macOS/Linux)

```bash
brew install pipeboard-co/tap/pipeboard
```

### Go install

```bash
go install github.com/pipeboard-co/pipeboard-cli@latest
```

### Binary download

Download the latest release from [GitHub Releases](https://github.com/pipeboard-co/pipeboard-cli/releases).

## Authentication

### API Token (recommended for automation)

```bash
export PIPEBOARD_API_TOKEN=<your-token>
```

Or store it in the config file:

```bash
pipeboard config set token <your-token>
```

Get your API token at [pipeboard.co/settings](https://pipeboard.co/settings).

### Browser OAuth (interactive)

```bash
pipeboard login
```

## Usage

Commands are auto-generated from Pipeboard's MCP tool definitions. Run `pipeboard --help` to see available commands.

```bash
# List Google Ads campaigns
pipeboard google-ads list-campaigns --customer-id 1234567890

# Get Meta Ads insights
pipeboard meta-ads get-insights --account-id act_123 --date-range last_30d

# Check version
pipeboard version
```

## Development

```bash
# Build
make build

# Run tests
make test

# Install locally
make install
```

## License

Apache License 2.0 — see [LICENSE](LICENSE).
