package cmd

import (
	"encoding/json"
	"testing"

	"github.com/pipeboard-co/pipeboard-cli/internal/client"
	"github.com/spf13/cobra"
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

// parseToolArgs registers the schema's flags on a throwaway command, parses
// argv through the same cobra machinery the real command uses, and returns the
// arguments object that would be sent to tools/call.
func parseToolArgs(t *testing.T, rawSchema string, argv ...string) map[string]interface{} {
	t.Helper()

	var schema InputSchema
	if err := json.Unmarshal([]byte(rawSchema), &schema); err != nil {
		t.Fatalf("unmarshal schema: %v", err)
	}

	cmd := &cobra.Command{Use: "test"}
	stringFlags, boolFlags := registerToolFlags(cmd, schema)

	if err := cmd.Flags().Parse(argv); err != nil {
		t.Fatalf("parse %v: %v", argv, err)
	}

	return buildToolArgs(cmd, schema, stringFlags, boolFlags)
}

// A boolean the caller explicitly set to false must reach the server as false.
// Sending only `true` made `--flag=false` indistinguishable from omitting the
// flag, so any tool param whose server-side default is true could never be
// turned off — delete_catalog_products could never leave dry run, and
// publish_facebook_page_post --published=false published a post regardless.
func TestBuildToolArgsBooleans(t *testing.T) {
	schema := `{
		"type": "object",
		"properties": {
			"catalog_id": {"type": "string"},
			"dry_run": {"type": "boolean"},
			"readback": {"type": "boolean"}
		}
	}`

	tests := []struct {
		name string
		argv []string
		want map[string]interface{}
	}{
		{
			name: "explicit false is sent, not dropped",
			argv: []string{"--catalog-id", "123", "--dry-run=false"},
			want: map[string]interface{}{"catalog_id": "123", "dry_run": false},
		},
		{
			name: "explicit true is sent",
			argv: []string{"--dry-run=true"},
			want: map[string]interface{}{"dry_run": true},
		},
		{
			name: "bare flag means true",
			argv: []string{"--dry-run"},
			want: map[string]interface{}{"dry_run": true},
		},
		{
			name: "unset booleans are omitted so the server default applies",
			argv: []string{"--catalog-id", "123"},
			want: map[string]interface{}{"catalog_id": "123"},
		},
		{
			name: "each boolean is tracked independently",
			argv: []string{"--dry-run=false"},
			want: map[string]interface{}{"dry_run": false},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parseToolArgs(t, schema, tc.argv...)
			if len(got) != len(tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
			for k, want := range tc.want {
				if got[k] != want {
					t.Errorf("arg %q = %v (%T), want %v (%T)", k, got[k], got[k], want, want)
				}
			}
		})
	}
}

// Underscored properties are exposed as hyphenated flags; the change-tracking
// lookup must use the same mapping or every boolean silently drops out again.
func TestBuildToolArgsHyphenatedBooleanFlag(t *testing.T) {
	schema := `{
		"type": "object",
		"properties": {"is_published": {"type": "boolean"}}
	}`

	got := parseToolArgs(t, schema, "--is-published=false")
	if v, ok := got["is_published"]; !ok || v != false {
		t.Errorf("is_published = %v (present=%v), want false", v, ok)
	}
}

// Non-boolean flags keep their schema-declared types.
func TestBuildToolArgsTypeCoercion(t *testing.T) {
	schema := `{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"limit": {"type": "integer"},
			"ratio": {"type": "number"},
			"product_ids": {"type": "array"},
			"filter": {"type": "object"}
		}
	}`

	got := parseToolArgs(t, schema,
		"--name", "hello",
		"--limit", "42",
		"--ratio", "1.5",
		"--product-ids", `["a","b"]`,
		"--filter", `{"k":"v"}`,
	)

	if got["name"] != "hello" {
		t.Errorf("name = %v, want hello", got["name"])
	}
	if got["limit"] != 42 {
		t.Errorf("limit = %v (%T), want 42 (int)", got["limit"], got["limit"])
	}
	if got["ratio"] != 1.5 {
		t.Errorf("ratio = %v (%T), want 1.5 (float64)", got["ratio"], got["ratio"])
	}
	ids, ok := got["product_ids"].([]interface{})
	if !ok || len(ids) != 2 || ids[0] != "a" {
		t.Errorf("product_ids = %v (%T), want [a b]", got["product_ids"], got["product_ids"])
	}
	filter, ok := got["filter"].(map[string]interface{})
	if !ok || filter["k"] != "v" {
		t.Errorf("filter = %v (%T), want map[k:v]", got["filter"], got["filter"])
	}
}
