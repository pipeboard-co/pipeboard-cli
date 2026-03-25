package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Version is set at build time via ldflags.
	Version = "dev"
	// Commit is set at build time via ldflags.
	Commit = "none"

	apiURL string
)

var rootCmd = &cobra.Command{
	Use:   "pipeboard",
	Short: "Manage ads across Meta, Google, and TikTok from your terminal",
	Long: `pipeboard is a command-line tool for managing advertising campaigns
across Meta Ads, Google Ads, and TikTok Ads.

All operations go through Pipeboard servers. Authenticate with:
  - PIPEBOARD_API_TOKEN environment variable
  - pipeboard login (browser-based OAuth)
  - pipeboard config set token <your-token>`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", getDefaultAPIURL(), "Pipeboard API base URL")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(debugCmd)
}

func getDefaultAPIURL() string {
	if url := os.Getenv("PIPEBOARD_API_URL"); url != "" {
		return url
	}
	return "https://mcp.pipeboard.co"
}

// getWebAPIURL returns the base URL for Pipeboard web API endpoints (not MCP).
// Debug and other REST endpoints live on the main domain, not on the MCP subdomain.
func getWebAPIURL() string {
	if url := os.Getenv("PIPEBOARD_WEB_URL"); url != "" {
		return url
	}
	return "https://pipeboard.co"
}

func getToken() (string, error) {
	// 1. Check environment variable
	if token := os.Getenv("PIPEBOARD_API_TOKEN"); token != "" {
		return token, nil
	}

	// 2. Check config file
	cfg, err := loadConfig()
	if err == nil && cfg.Token != "" {
		return cfg.Token, nil
	}

	return "", fmt.Errorf("not authenticated. Run 'pipeboard login' or set PIPEBOARD_API_TOKEN")
}
