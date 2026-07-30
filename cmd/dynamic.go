package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/pipeboard-co/pipeboard-cli/internal/client"
	"github.com/spf13/cobra"
)

const autoRefreshInterval = 1 * time.Hour

// registerDynamicCommands loads the tools cache and registers platform commands.
func registerDynamicCommands() {
	cache, err := loadToolsCache()
	if err != nil {
		return // No cache — user needs to run 'pipeboard refresh'
	}

	for _, server := range knownServers {
		serverCache, ok := cache.Servers[server.Path]
		if !ok || len(serverCache.Tools) == 0 {
			continue
		}

		platformCmd := &cobra.Command{
			Use:   server.CommandName,
			Short: server.Description,
		}

		for _, tool := range serverCache.Tools {
			cmd := buildToolCommand(server, tool)
			platformCmd.AddCommand(cmd)
		}

		rootCmd.AddCommand(platformCmd)
	}

	// Check for tool updates in the background
	maybeAutoRefresh(cache)
}

// maybeAutoRefresh checks whether the tools cache is stale and, if so,
// spawns a detached background process to check the hash and refresh if
// needed. The current command always uses the existing cache; updates
// take effect on the next invocation.
func maybeAutoRefresh(cache *ToolsCache) {
	// Guard against recursive spawning: the child sets this env var so
	// it won't spawn another child during its own init().
	if os.Getenv("PIPEBOARD_AUTO_REFRESH") == "1" {
		return
	}

	if time.Since(cache.LastCheckedAt) < autoRefreshInterval {
		return
	}

	// Spawn a detached "pipeboard refresh --if-changed" subprocess.
	// It survives parent exit and writes the cache for the next run.
	self, err := os.Executable()
	if err != nil {
		return
	}
	cmd := exec.Command(self, "refresh", "--if-changed")
	cmd.Env = append(os.Environ(), "PIPEBOARD_AUTO_REFRESH=1")
	setSysProcAttr(cmd)
	cmd.Stdout = nil
	cmd.Stderr = nil
	_ = cmd.Start()
}

// runAutoRefresh is the implementation behind "refresh --if-changed".
// It checks the hash first and only does a full refresh if tools changed.
func runAutoRefresh(baseURL, token string) error {
	cache, _ := loadToolsCache()

	hashResult, err := client.FetchToolsHash(baseURL, Version)
	if err != nil {
		return err
	}

	if cache != nil && hashResult.Hash == cache.Hash {
		// Tools unchanged — just bump the check timestamp
		cache.LastCheckedAt = time.Now()
		return saveToolsCache(cache)
	}

	// Hash changed or no cache — do a full refresh
	_, err = fetchAndCacheTools(baseURL, token, false)
	return err
}

// toolNameToCommandName converts an MCP tool name to a CLI command name.
// It strips the platform prefix and converts underscores to hyphens.
//
// Examples:
//
//	get_google_ads_campaigns  → get-campaigns
//	execute_google_ads_gaql_query → execute-gaql-query
//	get_campaigns             → get-campaigns
func toolNameToCommandName(toolName, prefix string) string {
	name := toolName
	if prefix != "" {
		name = strings.Replace(name, prefix, "", 1)
	}
	return strings.ReplaceAll(name, "_", "-")
}

// InputSchema is a simplified JSON Schema for MCP tool inputs.
type InputSchema struct {
	Type       string                    `json:"type"`
	Properties map[string]PropertySchema `json:"properties"`
	Required   []string                  `json:"required"`
}

// PropertySchema describes a single property in the input schema.
type PropertySchema struct {
	Type        interface{} `json:"type"` // string or []string (e.g. ["string", "null"])
	Description string      `json:"description"`
	Enum        []string    `json:"enum"`
}

// resolveType returns the primary type as a string, handling union types like ["string", "null"].
func (p PropertySchema) resolveType() string {
	switch t := p.Type.(type) {
	case string:
		return t
	case []interface{}:
		for _, v := range t {
			if s, ok := v.(string); ok && s != "null" {
				return s
			}
		}
	}
	return "string"
}

// propNameToFlagName maps a JSON Schema property name to its CLI flag name.
// Both flag registration and argument building go through this, so the two can
// never disagree about what a property's flag is called.
func propNameToFlagName(propName string) string {
	return strings.ReplaceAll(propName, "_", "-")
}

func buildToolCommand(server ServerConfig, tool client.ToolDefinition) *cobra.Command {
	cmdName := toolNameToCommandName(tool.Name, server.ToolPrefix)

	var schema InputSchema
	if len(tool.InputSchema) > 0 {
		_ = json.Unmarshal(tool.InputSchema, &schema)
	}

	cmd := &cobra.Command{
		Use:   cmdName,
		Short: truncate(tool.Description, 80),
		Long:  tool.Description,
	}

	stringFlags, boolFlags := registerToolFlags(cmd, schema)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return runDynamicTool(cmd, server, tool.Name, schema, stringFlags, boolFlags)
	}

	return cmd
}

// registerToolFlags declares one flag per JSON Schema property and returns the
// storage those flags parse into, keyed by property name.
func registerToolFlags(cmd *cobra.Command, schema InputSchema) (map[string]*string, map[string]*bool) {
	requiredSet := make(map[string]bool)
	for _, r := range schema.Required {
		requiredSet[r] = true
	}

	// Flag storage — one map per type we support
	stringFlags := make(map[string]*string)
	boolFlags := make(map[string]*bool)

	for propName, prop := range schema.Properties {
		flagName := propNameToFlagName(propName)
		desc := prop.Description
		if len(prop.Enum) > 0 {
			desc += " (options: " + strings.Join(prop.Enum, ", ") + ")"
		}

		switch prop.resolveType() {
		case "boolean":
			boolFlags[propName] = cmd.Flags().Bool(flagName, false, desc)
		default:
			// string, number, integer, array, object — all accepted as strings
			stringFlags[propName] = cmd.Flags().String(flagName, "", desc)
		}

		if requiredSet[propName] {
			cmd.MarkFlagRequired(flagName)
		}
	}

	return stringFlags, boolFlags
}

func runDynamicTool(cmd *cobra.Command, server ServerConfig, toolName string, schema InputSchema, stringFlags map[string]*string, boolFlags map[string]*bool) error {
	token, err := getToken()
	if err != nil {
		return err
	}

	c := client.New(apiURL, token, Version)

	fmt.Fprintf(os.Stderr, "Connecting to %s/%s...\n", apiURL, server.Path)
	if err := c.Initialize(server.Path); err != nil {
		return fmt.Errorf("handshake failed: %w", err)
	}

	toolArgs := buildToolArgs(cmd, schema, stringFlags, boolFlags)

	fmt.Fprintf(os.Stderr, "Calling %s...\n", toolName)
	result, err := c.CallTool(server.Path, toolName, toolArgs)
	if err != nil {
		return fmt.Errorf("tool call failed: %w", err)
	}

	return printToolResult(result)
}

// buildToolArgs turns parsed flag values into the tools/call arguments object,
// coercing each one to the type its JSON Schema declares.
func buildToolArgs(cmd *cobra.Command, schema InputSchema, stringFlags map[string]*string, boolFlags map[string]*bool) map[string]interface{} {
	toolArgs := make(map[string]interface{})

	for propName, val := range stringFlags {
		if val == nil || *val == "" {
			continue
		}
		prop := schema.Properties[propName]
		switch prop.resolveType() {
		case "integer":
			if n, err := strconv.Atoi(*val); err == nil {
				toolArgs[propName] = n
			} else {
				toolArgs[propName] = *val
			}
		case "number":
			if n, err := strconv.ParseFloat(*val, 64); err == nil {
				toolArgs[propName] = n
			} else {
				toolArgs[propName] = *val
			}
		case "object", "array":
			var parsed interface{}
			if err := json.Unmarshal([]byte(*val), &parsed); err == nil {
				toolArgs[propName] = parsed
			} else {
				toolArgs[propName] = *val
			}
		default:
			toolArgs[propName] = *val
		}
	}

	// Send a boolean whenever the caller actually typed the flag, including
	// `--flag=false`. Forwarding only `true` made an explicit false
	// indistinguishable from omitting the flag, so any tool param whose
	// server-side default is true could never be turned off — e.g.
	// `delete_catalog_products --dry-run=false` stayed in dry run and could
	// never execute, and `publish_facebook_page_post --published=false`
	// silently published a post the caller asked to keep as a draft.
	for propName, val := range boolFlags {
		if val == nil {
			continue
		}
		if cmd != nil && !cmd.Flags().Changed(propNameToFlagName(propName)) {
			continue
		}
		toolArgs[propName] = *val
	}

	return toolArgs
}

// printToolResult extracts text content from the MCP tools/call response and prints it.
func printToolResult(raw json.RawMessage) error {
	var toolResult struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		IsError bool `json:"isError"`
	}

	if err := json.Unmarshal(raw, &toolResult); err == nil && len(toolResult.Content) > 0 {
		for _, item := range toolResult.Content {
			if item.Type != "text" {
				continue
			}
			// Try to pretty-print as JSON; fall back to plain text
			var parsed interface{}
			if err := json.Unmarshal([]byte(item.Text), &parsed); err == nil {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				enc.Encode(parsed)
			} else {
				fmt.Println(item.Text)
			}
		}
		if toolResult.IsError {
			return fmt.Errorf("tool returned an error")
		}
		return nil
	}

	// Fallback: print raw result
	var parsed interface{}
	if err := json.Unmarshal(raw, &parsed); err == nil {
		return printJSON(parsed)
	}
	fmt.Println(string(raw))
	return nil
}
