# pipeboard-cli

Manage Meta Ads and Google Ads from your terminal. Single binary, zero dependencies, 50ms startup.

Built for AI coding agents (Claude Code, Cline, OpenClaw) and automation scripts that prefer `pipeboard meta-ads get-insights ...` over JSON-RPC. Every CLI invocation goes through [Pipeboard](https://pipeboard.co)'s cloud — your ad platform tokens never touch the local machine.

Commands are generated dynamically from Pipeboard's MCP tool definitions. When new tools ship server-side, they're available in the CLI automatically — no release required.

## Why use the CLI

- **All platforms, one binary** — Meta Ads and Google Ads without configuring separate MCP servers
- **More tools than local MCP** — 117 commands covering campaigns, creatives, audiences, reporting, and more
- **Fast for agents** — sub-50ms cold start. Agents shell out to `pipeboard` instead of implementing MCP JSON-RPC
- **Scriptable** — pipe JSON output into `jq`, chain commands in bash, run from CI

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

Download from [GitHub Releases](https://github.com/pipeboard-co/pipeboard-cli/releases). Builds for macOS (arm64/amd64), Linux (arm64/amd64), and Windows (amd64).

## Authentication

### API Token (recommended for automation)

```bash
export PIPEBOARD_API_TOKEN=<your-token>
```

Or store it in the config file:

```bash
pipeboard config set token <your-token>
```

Get your API token at [pipeboard.co/api-tokens](https://pipeboard.co/api-tokens).

### Browser OAuth (interactive)

```bash
pipeboard login
```

## Quick Start

```bash
# List Google Ads campaigns
pipeboard google-ads get-campaigns --customer-id 1234567890

# Get Meta Ads campaign insights
pipeboard meta-ads get-insights --object-id act_123 --date-preset last_30d

# Create a Meta Ads campaign
pipeboard meta-ads create-campaign --account-id act_123 --name "Spring Sale" --objective OUTCOME_SALES

# See all available commands
pipeboard google-ads --help
pipeboard meta-ads --help
```

Flags are generated from each tool's JSON Schema — required flags are enforced and enum values are shown in help text.

## Commands

### Google Ads

50 commands across campaigns, ad groups, ads, keywords, audiences, extensions, and reporting.

#### Campaigns

| Command | Description |
|---------|-------------|
| `create-campaign` | Create a new campaign |
| `get-campaigns` | List campaigns |
| `get-campaign-metrics` | Get campaign performance metrics with optional time-series |
| `update-campaign` | Update campaign name, status, or budget |
| `enable-campaign` | Enable a paused campaign |
| `pause-campaign` | Pause a campaign |
| `create-pmax-campaign` | Create a complete Performance Max campaign with assets |
| `get-pmax-asset-groups` | Get asset groups for a Performance Max campaign |
| `update-pmax-asset-group` | Add/remove assets and search themes in a PMax asset group |

#### Ad Groups

| Command | Description |
|---------|-------------|
| `create-ad-group` | Create a new ad group in a campaign |
| `get-ad-groups` | List ad groups |
| `get-ad-group-metrics` | Get ad group performance metrics |
| `update-ad-group` | Update ad group name, status, or bids |

#### Ads

| Command | Description |
|---------|-------------|
| `create-responsive-search-ad` | Create a Responsive Search Ad (RSA) |
| `get-ads` | List ads |
| `get-ad-metrics` | Get ad performance metrics |
| `enable-ad` | Enable a paused ad |
| `pause-ad` | Pause an ad |

#### Keywords

| Command | Description |
|---------|-------------|
| `add-keywords` | Add keywords to an ad group |
| `get-keywords` | List keywords |
| `get-keyword-metrics` | Get keyword performance metrics |
| `enable-keyword` | Enable paused keywords |
| `pause-keyword` | Pause keywords |
| `update-keyword-bid` | Update CPC bid for keywords |
| `remove-keywords` | Remove keywords from an ad group |
| `add-negative-keywords` | Add negative keywords to a campaign or ad group |
| `get-negative-keywords` | List negative keywords |
| `remove-negative-keywords` | Remove negative keywords |

#### Audiences & Targeting

| Command | Description |
|---------|-------------|
| `add-audience-to-campaign` | Add an audience to a campaign |
| `get-audiences` | List available audience segments |
| `get-device-performance` | Performance breakdown by device type |
| `get-geo-performance` | Performance breakdown by location |
| `get-hour-of-day-performance` | Performance breakdown by hour of day |
| `get-search-terms-report` | Actual search queries that triggered ads |
| `update-network-settings` | Update network targeting (Search, Display, etc.) |

#### Extensions

| Command | Description |
|---------|-------------|
| `create-callout` | Create a callout extension |
| `create-sitelink` | Create a sitelink extension |
| `create-structured-snippet` | Create a structured snippet extension |
| `get-extensions` | List ad extensions for campaigns |
| `remove-extension` | Remove an extension from a campaign |
| `update-extension-status` | Enable or pause an extension |

#### Assets

| Command | Description |
|---------|-------------|
| `list-assets` | List creative assets in an account |
| `upload-asset` | Upload an image, text, or YouTube video asset |

#### Account & Reporting

| Command | Description |
|---------|-------------|
| `get-account-info` | Get account details |
| `list-customers` | List accessible customer accounts |
| `execute-gaql-query` | Run a raw GAQL query |
| `execute-mutate` | Run a raw mutate operation |
| `query-api-docs` | Search Google Ads API documentation |
| `create-email-report` | Send an AI-powered email performance report |
| `submit-feedback` | Report a bug or request a feature |

---

### Meta Ads

67 commands across campaigns, ad sets, ads, creatives, audiences, lead gen, and reporting.

#### Campaigns

| Command | Description |
|---------|-------------|
| `create-campaign` | Create a new campaign |
| `get-campaigns` | List campaigns |
| `get-campaign-details` | Get detailed campaign info |
| `update-campaign` | Update campaign settings |
| `duplicate-campaign` | Duplicate a campaign with all ad sets and ads |
| `bulk-update-campaigns` | Update multiple campaigns at once |
| `create-budget-schedule` | Schedule budget changes for a campaign |

#### Ad Sets

| Command | Description |
|---------|-------------|
| `create-adset` | Create a new ad set |
| `get-adsets` | List ad sets |
| `get-adset-details` | Get detailed ad set info |
| `update-adset` | Update ad set settings, budgets, and frequency caps |
| `duplicate-adset` | Duplicate an ad set with its ads |
| `bulk-update-adsets` | Update multiple ad sets at once |

#### Ads

| Command | Description |
|---------|-------------|
| `create-ad` | Create an ad with an existing creative |
| `get-ads` | List ads |
| `get-ad-details` | Get detailed ad info |
| `update-ad` | Update ad settings |
| `duplicate-ad` | Duplicate an ad |
| `bulk-update-ads` | Update multiple ads at once |

#### Creatives

| Command | Description |
|---------|-------------|
| `create-ad-creative` | Create a creative from an image or video |
| `create-carousel-ad-creative` | Create a carousel creative with multiple cards |
| `get-ad-creatives` | Get creative details for an ad |
| `get-creative-details` | Get creative details by creative ID |
| `get-ad-previews` | Get an ad or creative preview (iframe HTML) |
| `update-ad-creative` | Update creative name or optimization settings |
| `duplicate-creative` | Duplicate a creative |
| `bulk-create-ad-creatives` | Create multiple creatives in a batch |
| `bulk-get-ad-creatives` | Fetch creative details for multiple ads |

#### Media

| Command | Description |
|---------|-------------|
| `upload-ad-image` | Upload an image for use in creatives |
| `get-ad-image` | Download and visualize an ad image |
| `upload-ad-video` | Upload a video for use in creatives |
| `get-ad-video` | Get video details and source URL |
| `bulk-upload-ad-images` | Upload multiple images at once |
| `bulk-upload-ad-videos` | Upload multiple videos at once |

#### Audiences & Targeting

| Command | Description |
|---------|-------------|
| `create-custom-audience` | Create a custom audience |
| `create-lookalike-audience` | Create a lookalike audience from a seed |
| `get-custom-audiences` | List custom audiences |
| `delete-custom-audience` | Delete a custom audience |
| `add-users-to-audience` | Upload a customer list to an audience |
| `estimate-audience-size` | Estimate reach for targeting specs |
| `search-interests` | Search interest targeting options |
| `get-interest-suggestions` | Get suggestions based on existing interests |
| `bulk-search-interests` | Search interests across multiple keywords |
| `search-behaviors` | List behavior targeting options |
| `search-demographics` | List demographic targeting options |
| `search-geo-locations` | Search geographic targeting locations |

#### Insights & Reporting

| Command | Description |
|---------|-------------|
| `get-insights` | Get performance insights for any object |
| `bulk-get-insights` | Get insights for multiple objects at once |
| `create-email-report` | Send an AI-powered email performance report |
| `list-email-reports` | List email report schedules |
| `update-email-report` | Update an email report schedule |
| `delete-email-report` | Delete an email report schedule |

#### Account & Pages

| Command | Description |
|---------|-------------|
| `get-ad-accounts` | List accessible ad accounts |
| `get-account-info` | Get account details |
| `get-account-pages` | Get pages associated with an account |
| `get-instagram-accounts` | Get Instagram accounts for an ad account |
| `get-pixels` | List Meta pixels for an account |
| `search-pages-by-name` | Search for pages by name |

#### Lead Generation

| Command | Description |
|---------|-------------|
| `create-lead-gen-form` | Create a lead gen form on a Page |
| `get-lead-gen-forms` | List lead gen forms for a Page |
| `publish-lead-gen-draft-form` | Publish a draft lead gen form |
| `update-lead-gen-form-status` | Archive or reactivate a form |

#### Product Catalogs

| Command | Description |
|---------|-------------|
| `list-catalogs` | List product catalogs for an account |
| `list-product-sets` | List product sets in a catalog |

#### Other

| Command | Description |
|---------|-------------|
| `submit-feedback` | Report a bug or request a feature |

## How It Works

```
pipeboard <platform> <command> [flags]
    │
    ▼
MCP JSON-RPC client (HTTP POST)
    │
    ▼
https://mcp.pipeboard.co/<server>
    │
    ▼
Meta / Google APIs
```

The CLI is a thin client. All API calls go through Pipeboard's cloud — auth tokens, rate limiting, and logging are handled server-side. This is why the CLI repo contains no secrets and can be fully open-source.

## Also Available: MCP Servers

If you prefer MCP over CLI, Pipeboard offers remote MCP servers for the same platforms:

| Platform | Remote MCP URL |
|---|---|
| Meta Ads | `https://mcp.pipeboard.co/meta-ads-mcp` |
| Google Ads | `https://mcp.pipeboard.co/google-ads-mcp` |

Works with Claude Pro/Max, Cursor, and any MCP client. Set up at [pipeboard.co](https://pipeboard.co).

## Community

- [Discord](https://discord.gg/YzMwQ8zrjr) — join the community
- [Email](mailto:info@pipeboard.co) — support and questions
- [Issues](https://github.com/pipeboard-co/pipeboard-cli/issues) — bug reports and feature requests

## Development

```bash
make build     # Build binary to bin/pipeboard
make test      # Run tests
make install   # Install to $GOPATH/bin
make lint      # Run go vet
```

## License

Apache License 2.0 — see [LICENSE](LICENSE).
