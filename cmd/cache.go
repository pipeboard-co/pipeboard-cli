package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/pipeboard-co/pipeboard-cli/internal/client"
)

// ToolsCache stores tool definitions fetched from MCP servers.
type ToolsCache struct {
	Servers       map[string]ServerToolsCache `json:"servers"`
	UpdatedAt     time.Time                   `json:"updated_at"`
	Hash          string                      `json:"hash,omitempty"`
	LastCheckedAt time.Time                   `json:"last_checked_at,omitempty"`
}

// ServerToolsCache holds the tools for a single MCP server.
type ServerToolsCache struct {
	Tools []client.ToolDefinition `json:"tools"`
}

func toolsCachePath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "tools-cache.json"), nil
}

func loadToolsCache() (*ToolsCache, error) {
	path, err := toolsCachePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cache ToolsCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}
	return &cache, nil
}

func saveToolsCache(cache *ToolsCache) error {
	path, err := toolsCachePath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
