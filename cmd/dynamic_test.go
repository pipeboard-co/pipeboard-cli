package cmd

import (
	"encoding/json"
	"testing"

	"github.com/pipeboard-co/pipeboard-cli/internal/client"
)

func TestToolNameToCommandName(t *testing.T) {
	tests := []struct {
		toolName string
		prefix   string
		want     string
	}{
		{"get_google_ads_campaigns", "google_ads_", "get-campaigns"},
		{"execute_google_ads_gaql_query", "google_ads_", "execute-gaql-query"},
		{"add_google_ads_keywords", "google_ads_", "add-keywords"},
		{"create_google_ads_responsive_search_ad", "google_ads_", "create-responsive-search-ad"},
		{"list_google_ads_customers", "google_ads_", "list-customers"},
		{"submit_feedback", "google_ads_", "submit-feedback"},
		// Meta Ads — no prefix
		{"get_campaigns", "", "get-campaigns"},
		{"get_ad_creatives", "", "get-ad-creatives"},
		{"create_ad", "", "create-ad"},
		{"bulk_upload_ad_images", "", "bulk-upload-ad-images"},
	}

	for _, tt := range tests {
		t.Run(tt.toolName, func(t *testing.T) {
			got := toolNameToCommandName(tt.toolName, tt.prefix)
			if got != tt.want {
				t.Errorf("toolNameToCommandName(%q, %q) = %q, want %q", tt.toolName, tt.prefix, got, tt.want)
			}
		})
	}
}

func TestBuildToolCommand(t *testing.T) {
	server := ServerConfig{
		Path:        "google-ads-mcp",
		CommandName: "google-ads",
		ToolPrefix:  "google_ads_",
		Description: "Test",
	}

	schema := `{
		"type": "object",
		"properties": {
			"customer_id": {"type": "string", "description": "The customer ID"},
			"limit": {"type": "integer", "description": "Max results"},
			"include_drafts": {"type": "boolean", "description": "Include drafts"},
			"status": {"type": "string", "description": "Filter by status", "enum": ["ENABLED", "PAUSED"]}
		},
		"required": ["customer_id"]
	}`

	tool := client.ToolDefinition{
		Name:        "get_google_ads_campaigns",
		Description: "Get campaigns for a Google Ads account",
		InputSchema: json.RawMessage(schema),
	}

	cmd := buildToolCommand(server, tool)

	if cmd.Use != "get-campaigns" {
		t.Errorf("Use = %q, want %q", cmd.Use, "get-campaigns")
	}

	// Check flags exist
	for _, flagName := range []string{"customer-id", "limit", "include-drafts", "status"} {
		if cmd.Flags().Lookup(flagName) == nil {
			t.Errorf("expected flag --%s to exist", flagName)
		}
	}

	// Check required flag
	f := cmd.Flags().Lookup("customer-id")
	if f == nil {
		t.Fatal("customer-id flag not found")
	}
	// Cobra marks required flags via annotations
	ann := f.Annotations
	if ann == nil {
		t.Error("expected customer-id to have required annotation")
	}

	// Check enum in description
	statusFlag := cmd.Flags().Lookup("status")
	if statusFlag == nil {
		t.Fatal("status flag not found")
	}
	if want := "(options: ENABLED, PAUSED)"; !contains(statusFlag.Usage, want) {
		t.Errorf("status usage = %q, want to contain %q", statusFlag.Usage, want)
	}
}

func TestBuildToolCommandUnionType(t *testing.T) {
	server := ServerConfig{
		Path:        "meta-ads-mcp",
		CommandName: "meta-ads",
	}

	// Some schemas use union types like ["string", "null"]
	schema := `{
		"type": "object",
		"properties": {
			"account_id": {"type": "string", "description": "The account ID"},
			"fields": {"type": ["string", "null"], "description": "Fields to return"}
		},
		"required": ["account_id"]
	}`

	tool := client.ToolDefinition{
		Name:        "get_campaigns",
		Description: "Get campaigns",
		InputSchema: json.RawMessage(schema),
	}

	cmd := buildToolCommand(server, tool)

	if cmd.Use != "get-campaigns" {
		t.Errorf("Use = %q, want %q", cmd.Use, "get-campaigns")
	}

	// Union type should resolve to string flag
	f := cmd.Flags().Lookup("fields")
	if f == nil {
		t.Fatal("fields flag not found")
	}
	if f.Value.Type() != "string" {
		t.Errorf("fields type = %q, want string", f.Value.Type())
	}
}

func TestPropertySchemaResolveType(t *testing.T) {
	tests := []struct {
		name string
		typ  interface{}
		want string
	}{
		{"simple string", "string", "string"},
		{"simple integer", "integer", "integer"},
		{"simple boolean", "boolean", "boolean"},
		{"union nullable", []interface{}{"string", "null"}, "string"},
		{"union number", []interface{}{"null", "number"}, "number"},
		{"nil type", nil, "string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := PropertySchema{Type: tt.typ}
			got := p.resolveType()
			if got != tt.want {
				t.Errorf("resolveType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
