package cmd

// ServerConfig defines a known Pipeboard MCP server.
type ServerConfig struct {
	Path        string // MCP server path, e.g. "google-ads-mcp"
	CommandName string // CLI subcommand name, e.g. "google-ads"
	ToolPrefix  string // Prefix to strip from tool names, e.g. "google_ads_"
	Description string // Short description for help text
}

// knownServers is the registry of MCP servers the CLI can talk to.
var knownServers = []ServerConfig{
	{
		Path:        "google-ads-mcp",
		CommandName: "google-ads",
		ToolPrefix:  "google_ads_",
		Description: "Manage Google Ads campaigns, ad groups, keywords, and more",
	},
	{
		Path:        "meta-ads-mcp",
		CommandName: "meta-ads",
		ToolPrefix:  "",
		Description: "Manage Meta Ads campaigns, ad sets, ads, and creatives",
	},
}
