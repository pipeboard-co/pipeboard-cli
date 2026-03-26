package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/pipeboard-co/pipeboard-cli/internal/client"
	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Interact with Pipeboard MCP servers via JSON-RPC",
}

var (
	mcpServer string
)

var mcpToolsListCmd = &cobra.Command{
	Use:   "tools-list",
	Short: "List available tools from an MCP server",
	RunE:  runMCPToolsList,
}

var (
	mcpToolsListFilter  string
	mcpToolsListVerbose bool
)

var mcpToolsCallCmd = &cobra.Command{
	Use:   "tools-call",
	Short: "Call a tool on an MCP server",
	RunE:  runMCPToolsCall,
}

var (
	mcpToolName string
	mcpToolArgs string
)

func init() {
	mcpCmd.PersistentFlags().StringVar(&mcpServer, "server", "", "MCP server path, e.g. meta-ads-mcp (required)")
	mcpCmd.MarkPersistentFlagRequired("server")

	mcpToolsCallCmd.Flags().StringVar(&mcpToolName, "tool", "", "Tool name to call (required)")
	mcpToolsCallCmd.Flags().StringVar(&mcpToolArgs, "args", "{}", "JSON object of tool arguments")
	mcpToolsCallCmd.MarkFlagRequired("tool")

	mcpToolsListCmd.Flags().StringVar(&mcpToolsListFilter, "filter", "", "Filter tools by name substring")
	mcpToolsListCmd.Flags().BoolVar(&mcpToolsListVerbose, "verbose", false, "Show full tool descriptions and schemas")

	mcpCmd.AddCommand(mcpToolsListCmd)
	mcpCmd.AddCommand(mcpToolsCallCmd)
}

func newMCPClient() (*client.Client, error) {
	token, err := getToken()
	if err != nil {
		return nil, err
	}
	return client.New(apiURL, token), nil
}

func runMCPToolsList(cmd *cobra.Command, args []string) error {
	c, err := newMCPClient()
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Connecting to %s/%s...\n", apiURL, mcpServer)

	if err := c.Initialize(mcpServer); err != nil {
		return fmt.Errorf("handshake failed: %w", err)
	}

	result, err := c.ListTools(mcpServer)
	if err != nil {
		return fmt.Errorf("listing tools: %w", err)
	}

	// Filter if requested
	tools := result.Tools
	if mcpToolsListFilter != "" {
		var filtered []client.ToolDefinition
		for _, t := range tools {
			if strings.Contains(strings.ToLower(t.Name), strings.ToLower(mcpToolsListFilter)) {
				filtered = append(filtered, t)
			}
		}
		tools = filtered
	}

	fmt.Fprintf(os.Stderr, "Found %d tools (showing %d)\n\n", len(result.Tools), len(tools))

	if mcpToolsListVerbose {
		for _, tool := range tools {
			fmt.Printf("=== %s ===\n", tool.Name)
			fmt.Printf("Description:\n%s\n", tool.Description)
			if len(tool.InputSchema) > 0 {
				var schema interface{}
				if err := json.Unmarshal(tool.InputSchema, &schema); err == nil {
					schemaJSON, _ := json.MarshalIndent(schema, "", "  ")
					fmt.Printf("Schema:\n%s\n", string(schemaJSON))
				}
			}
			fmt.Println()
		}
	} else {
		for _, tool := range tools {
			fmt.Printf("%-40s %s\n", tool.Name, truncate(tool.Description, 80))
		}
	}

	return nil
}

func runMCPToolsCall(cmd *cobra.Command, args []string) error {
	c, err := newMCPClient()
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Connecting to %s/%s...\n", apiURL, mcpServer)

	if err := c.Initialize(mcpServer); err != nil {
		return fmt.Errorf("handshake failed: %w", err)
	}

	var toolArgs map[string]interface{}
	if err := json.Unmarshal([]byte(mcpToolArgs), &toolArgs); err != nil {
		return fmt.Errorf("invalid JSON in --args: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Calling %s...\n", mcpToolName)

	result, err := c.CallTool(mcpServer, mcpToolName, toolArgs)
	if err != nil {
		return fmt.Errorf("tool call failed: %w", err)
	}

	// Pretty-print the result
	var parsed interface{}
	if err := json.Unmarshal(result, &parsed); err != nil {
		// Not JSON, print raw
		fmt.Println(string(result))
		return nil
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(parsed)
}

func truncate(s string, max int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
