package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/pipeboard-co/pipeboard-cli/internal/client"
	"github.com/spf13/cobra"
)

var refreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Fetch and cache tool definitions from Pipeboard MCP servers",
	Long: `Fetches available tools from all Pipeboard MCP servers and caches them locally.
This enables platform commands like 'pipeboard google-ads get-campaigns'.

Run this after first install or when new tools become available.`,
	RunE: runRefresh,
}

func runRefresh(cmd *cobra.Command, args []string) error {
	token, err := getToken()
	if err != nil {
		return err
	}

	c := client.New(apiURL, token)
	cache := &ToolsCache{
		Servers:   make(map[string]ServerToolsCache),
		UpdatedAt: time.Now(),
	}

	totalTools := 0
	for _, server := range knownServers {
		fmt.Fprintf(os.Stderr, "Fetching tools from %s... ", server.Path)

		if err := c.Initialize(server.Path); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			continue
		}

		result, err := c.ListTools(server.Path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			continue
		}

		cache.Servers[server.Path] = ServerToolsCache{Tools: result.Tools}
		totalTools += len(result.Tools)
		fmt.Fprintf(os.Stderr, "%d tools\n", len(result.Tools))
	}

	if err := saveToolsCache(cache); err != nil {
		return fmt.Errorf("saving cache: %w", err)
	}

	path, _ := toolsCachePath()
	fmt.Fprintf(os.Stderr, "\nCached %d tools to %s\n", totalTools, path)
	return nil
}
