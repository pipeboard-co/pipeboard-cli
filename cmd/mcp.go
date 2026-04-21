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
	mcpJSON   bool
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
	mcpCmd.PersistentFlags().BoolVar(&mcpJSON, "json", false, "Emit a structured JSON envelope on stdout and suppress status messages on stderr. Unwraps the MCP content wrapper and parses text payloads as JSON when possible.")

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
	return client.New(apiURL, token, Version), nil
}

func runMCPToolsList(cmd *cobra.Command, args []string) error {
	if mcpJSON {
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true
	}

	c, err := newMCPClient()
	if err != nil {
		return emitJSONError(err)
	}

	if !mcpJSON {
		fmt.Fprintf(os.Stderr, "Connecting to %s/%s...\n", apiURL, mcpServer)
	}

	if err := c.Initialize(mcpServer); err != nil {
		return emitJSONError(fmt.Errorf("handshake failed: %w", err))
	}

	result, err := c.ListTools(mcpServer)
	if err != nil {
		return emitJSONError(fmt.Errorf("listing tools: %w", err))
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

	if mcpJSON {
		return emitJSON(map[string]interface{}{
			"ok":    true,
			"tools": tools,
			"total": len(result.Tools),
		})
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
	if mcpJSON {
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true
	}

	c, err := newMCPClient()
	if err != nil {
		return emitJSONError(err)
	}

	if !mcpJSON {
		fmt.Fprintf(os.Stderr, "Connecting to %s/%s...\n", apiURL, mcpServer)
	}

	if err := c.Initialize(mcpServer); err != nil {
		return emitJSONError(fmt.Errorf("handshake failed: %w", err))
	}

	var toolArgs map[string]interface{}
	if err := json.Unmarshal([]byte(mcpToolArgs), &toolArgs); err != nil {
		return emitJSONError(fmt.Errorf("invalid JSON in --args: %w", err))
	}

	if !mcpJSON {
		fmt.Fprintf(os.Stderr, "Calling %s...\n", mcpToolName)
	}

	result, err := c.CallTool(mcpServer, mcpToolName, toolArgs)
	if err != nil {
		return emitJSONError(fmt.Errorf("tool call failed: %w", err))
	}

	if mcpJSON {
		return emitToolResultJSON(result)
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

// emitJSON writes a structured JSON envelope to stdout with pretty indentation.
func emitJSON(payload map[string]interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(payload)
}

// emitJSONError emits a structured error on stdout in JSON mode, then returns
// the original error so the command still exits non-zero. In non-JSON mode it
// just returns the error (cobra prints it to stderr).
func emitJSONError(err error) error {
	if !mcpJSON {
		return err
	}
	_ = emitJSON(map[string]interface{}{
		"ok":    false,
		"error": err.Error(),
	})
	return err
}

// emitToolResultJSON unwraps the MCP tools/call envelope (content[].text +
// isError) into a flatter shape that callers can consume without a second
// parse step. If content[0].text is valid JSON, it's parsed into data;
// otherwise data holds the raw text.
func emitToolResultJSON(raw json.RawMessage) error {
	var env struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		IsError bool `json:"isError"`
	}

	out := map[string]interface{}{"ok": true}

	if err := json.Unmarshal(raw, &env); err != nil || len(env.Content) == 0 {
		// Fallback: return the raw result under data.
		var parsed interface{}
		if json.Unmarshal(raw, &parsed) != nil {
			parsed = string(raw)
		}
		out["data"] = parsed
		return emitJSON(out)
	}

	text := env.Content[0].Text
	var parsed interface{}
	if json.Unmarshal([]byte(text), &parsed) == nil {
		out["data"] = parsed
	} else {
		out["data"] = text
	}

	if env.IsError {
		out["ok"] = false
		if s, ok := out["data"].(string); ok {
			out["error"] = s
		} else {
			out["error"] = "tool returned isError=true"
		}
		if err := emitJSON(out); err != nil {
			return err
		}
		return fmt.Errorf("tool reported error")
	}

	return emitJSON(out)
}

func truncate(s string, max int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
