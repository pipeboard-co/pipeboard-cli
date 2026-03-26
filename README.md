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

## Quick Start

After authenticating, fetch the available tools:

```bash
pipeboard refresh
```

This downloads tool definitions from Pipeboard's MCP servers and caches them locally at `~/.pipeboard/tools-cache.json`. Run this once after install, or again when new tools are available.

## Usage

Commands are auto-generated from Pipeboard's MCP tool definitions. Each ad platform is a top-level command with subcommands for every available operation.

```bash
# List Google Ads campaigns
pipeboard google-ads get-campaigns --customer-id 1234567890

# Execute a raw GAQL query
pipeboard google-ads execute-gaql-query --customer-id 1234567890 --query "SELECT campaign.id, campaign.name FROM campaign LIMIT 10"

# Get Meta Ads campaign insights
pipeboard meta-ads get-insights --object-id act_123 --date-preset last_30d

# Create a Meta Ads campaign
pipeboard meta-ads create-campaign --account-id act_123 --name "Spring Sale" --objective OUTCOME_SALES

# See all available commands for a platform
pipeboard google-ads --help
pipeboard meta-ads --help

# Check version
pipeboard version
```

Flags are generated from each tool's JSON Schema — required flags are enforced and enum values are shown in help text.

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
