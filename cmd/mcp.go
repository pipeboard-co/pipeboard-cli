package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
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

	if isToolNotFound(result) {
		if !mcpJSON {
			fmt.Fprintf(os.Stderr, "Tool not found; refreshing tool list and retrying...\n")
		}
		tools, listErr := c.ListTools(mcpServer)
		if listErr != nil {
			return emitJSONError(fmt.Errorf("tool not found and tools/list failed: %w", listErr))
		}
		if toolExists(tools.Tools, mcpToolName) {
			result, err = c.CallTool(mcpServer, mcpToolName, toolArgs)
			if err != nil {
				return emitJSONError(fmt.Errorf("tool call failed: %w", err))
			}
		} else {
			msg := fmt.Sprintf("tool %q not found on server %q", mcpToolName, mcpServer)
			if suggestions := similarTools(tools.Tools, mcpToolName, 5); len(suggestions) > 0 {
				msg += " (did you mean: " + strings.Join(suggestions, ", ") + "?)"
			}
			return emitJSONError(fmt.Errorf("%s", msg))
		}
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

// isToolNotFound reports whether a tools/call result is an isError envelope
// whose text begins with "Tool not found". That signal is what triggers an
// implicit tools/list refresh and one retry.
func isToolNotFound(raw json.RawMessage) bool {
	var env struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
		IsError bool `json:"isError"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		return false
	}
	if !env.IsError || len(env.Content) == 0 {
		return false
	}
	return strings.HasPrefix(env.Content[0].Text, "Tool not found")
}

func toolExists(tools []client.ToolDefinition, name string) bool {
	for _, t := range tools {
		if t.Name == name {
			return true
		}
	}
	return false
}

// similarTools returns up to max tool names that share underscore-separated
// tokens with name, ranked by shared-token count. Used to produce a
// "did you mean" hint when a tool is genuinely missing.
func similarTools(tools []client.ToolDefinition, name string, max int) []string {
	tokens := strings.Split(strings.ToLower(name), "_")
	type scored struct {
		name  string
		score int
	}
	var candidates []scored
	for _, t := range tools {
		lower := strings.ToLower(t.Name)
		score := 0
		for _, tok := range tokens {
			if len(tok) > 1 && strings.Contains(lower, tok) {
				score++
			}
		}
		if score > 0 {
			candidates = append(candidates, scored{t.Name, score})
		}
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})
	if len(candidates) > max {
		candidates = candidates[:max]
	}
	out := make([]string, len(candidates))
	for i, c := range candidates {
		out[i] = c.name
	}
	return out
}

func truncate(s string, max int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
