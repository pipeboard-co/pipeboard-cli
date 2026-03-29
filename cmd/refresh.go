package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/pipeboard-co/pipeboard-cli/internal/client"
	"github.com/spf13/cobra"
)

var ifChanged bool

var refreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Fetch and cache tool definitions from Pipeboard MCP servers",
	Long: `Fetches available tools from all Pipeboard MCP servers and caches them locally.
This enables platform commands like 'pipeboard google-ads get-campaigns'.

Normally you don't need to run this — the CLI automatically checks for
updates once per hour and refreshes in the background. Use this command
to force an immediate refresh.`,
	RunE: runRefresh,
}

func init() {
	refreshCmd.Flags().BoolVar(&ifChanged, "if-changed", false, "Only refresh if tools hash has changed (used by auto-refresh)")
	refreshCmd.Flags().MarkHidden("if-changed")
}

func runRefresh(cmd *cobra.Command, args []string) error {
	token, err := getToken()
	if err != nil {
		return err
	}

	if ifChanged {
		return runAutoRefresh(apiURL, token)
	}

	cache, err := fetchAndCacheTools(apiURL, token, true)
	if err != nil {
		return err
	}

	totalTools := 0
	for _, sc := range cache.Servers {
		totalTools += len(sc.Tools)
	}
	path, _ := toolsCachePath()
	fmt.Fprintf(os.Stderr, "\nCached %d tools to %s\n", totalTools, path)
	return nil
}

// fetchAndCacheTools fetches tools from all servers and saves the cache.
// If verbose is true, progress is printed to stderr.
func fetchAndCacheTools(baseURL, token string, verbose bool) (*ToolsCache, error) {
	c := client.New(baseURL, token, Version)
	cache := &ToolsCache{
		Servers:       make(map[string]ServerToolsCache),
		UpdatedAt:     time.Now(),
		LastCheckedAt: time.Now(),
	}

	for _, server := range knownServers {
		if verbose {
			fmt.Fprintf(os.Stderr, "Fetching tools from %s... ", server.Path)
		}

		if err := c.Initialize(server.Path); err != nil {
			if verbose {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
			}
			continue
		}

		result, err := c.ListTools(server.Path)
		if err != nil {
			if verbose {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
			}
			continue
		}

		cache.Servers[server.Path] = ServerToolsCache{Tools: result.Tools}
		if verbose {
			fmt.Fprintf(os.Stderr, "%d tools\n", len(result.Tools))
		}
	}

	// Fetch and store the current hash
	if hashResult, err := client.FetchToolsHash(baseURL); err == nil {
		cache.Hash = hashResult.Hash
	}

	if err := saveToolsCache(cache); err != nil {
		return nil, fmt.Errorf("saving cache: %w", err)
	}
	return cache, nil
}
