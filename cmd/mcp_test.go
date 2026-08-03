package cmd

import (
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/pipeboard-co/pipeboard-cli/internal/client"
)

// captureStdout runs fn with os.Stdout redirected to a pipe and returns the
// captured bytes. Used to test emitJSON-family helpers without plumbing a
// writer through every caller.
func captureStdout(t *testing.T, fn func()) []byte {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	done := make(chan []byte, 1)
	go func() {
		b, _ := io.ReadAll(r)
		done <- b
	}()

	fn()
	w.Close()
	return <-done
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		max      int
		expected string
	}{
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"this is a longer string", 10, "this is..."},
		{"has\nnewlines\nin it", 80, "has newlines in it"},
		{"", 5, ""},
	}

	for _, tt := range tests {
		got := truncate(tt.input, tt.max)
		if got != tt.expected {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.expected)
		}
	}
}

func TestEmitToolResultJSON_UnwrapsJSONTextContent(t *testing.T) {
	raw := json.RawMessage(`{"content":[{"type":"text","text":"{\"name\":\"ad-1\",\"id\":42}"}],"isError":false}`)

	out := captureStdout(t, func() {
		if err := emitToolResultJSON(raw); err != nil {
			t.Fatalf("emitToolResultJSON: %v", err)
		}
	})

	var got map[string]interface{}
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("stdout not valid JSON: %v\n%s", err, out)
	}
	if got["ok"] != true {
		t.Errorf("ok = %v, want true", got["ok"])
	}
	data, ok := got["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("data not an object: %T", got["data"])
	}
	if data["name"] != "ad-1" {
		t.Errorf("data.name = %v, want ad-1", data["name"])
	}
}

func TestEmitToolResultJSON_IsErrorSetsOkFalse(t *testing.T) {
	raw := json.RawMessage(`{"content":[{"type":"text","text":"auth expired"}],"isError":true}`)

	var retErr error
	out := captureStdout(t, func() {
		retErr = emitToolResultJSON(raw)
	})
	if retErr == nil {
		t.Fatalf("expected error return so exit code is non-zero")
	}

	var got map[string]interface{}
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("stdout not valid JSON: %v\n%s", err, out)
	}
	if got["ok"] != false {
		t.Errorf("ok = %v, want false", got["ok"])
	}
	if got["error"] != "auth expired" {
		t.Errorf("error = %v, want %q", got["error"], "auth expired")
	}
	if got["data"] != "auth expired" {
		t.Errorf("data = %v, want raw text", got["data"])
	}
}

func TestEmitToolResultJSON_PassthroughWhenNotEnvelope(t *testing.T) {
	raw := json.RawMessage(`{"foo":"bar"}`)

	out := captureStdout(t, func() {
		if err := emitToolResultJSON(raw); err != nil {
			t.Fatalf("emitToolResultJSON: %v", err)
		}
	})

	var got map[string]interface{}
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("stdout not valid JSON: %v\n%s", err, out)
	}
	data, ok := got["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("data not an object: %T", got["data"])
	}
	if data["foo"] != "bar" {
		t.Errorf("data.foo = %v, want bar", data["foo"])
	}
}

func TestIsToolNotFound(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want bool
	}{
		{"tool not found envelope", `{"content":[{"type":"text","text":"Tool not found: foo"}],"isError":true}`, true},
		{"isError but different message", `{"content":[{"type":"text","text":"auth expired"}],"isError":true}`, false},
		{"not an error", `{"content":[{"type":"text","text":"Tool not found: foo"}],"isError":false}`, false},
		{"no content", `{"content":[],"isError":true}`, false},
		{"garbage", `not json`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isToolNotFound(json.RawMessage(tt.raw)); got != tt.want {
				t.Errorf("isToolNotFound = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSimilarTools(t *testing.T) {
	tools := []client.ToolDefinition{
		{Name: "list_snap_audiences"},
		{Name: "list_snap_campaigns"},
		{Name: "list_meta_campaigns"},
		{Name: "get_snap_segment_details"},
	}
	got := similarTools(tools, "list_snap_segments", 3)
	if len(got) == 0 {
		t.Fatalf("expected suggestions, got none")
	}
	if got[0] != "list_snap_audiences" && got[0] != "list_snap_campaigns" {
		t.Errorf("top suggestion should be a list_snap_* tool, got %q", got[0])
	}
}

func TestToolCategoryLabel(t *testing.T) {
	tr := true
	fa := false

	tests := []struct {
		name string
		a    *client.ToolAnnotations
		want string
	}{
		{"nil annotations -> ?", nil, "?"},
		{"readOnly", &client.ToolAnnotations{ReadOnlyHint: &tr}, "read"},
		{"destructive (readOnly false)", &client.ToolAnnotations{
			ReadOnlyHint: &fa, DestructiveHint: &tr, IdempotentHint: &fa,
		}, "del"},
		{"idempotent mutation", &client.ToolAnnotations{
			ReadOnlyHint: &fa, DestructiveHint: &fa, IdempotentHint: &tr,
		}, "idem"},
		{"additive write", &client.ToolAnnotations{
			ReadOnlyHint: &fa, DestructiveHint: &fa, IdempotentHint: &fa,
		}, "write"},
		{"empty annotations object (all hints absent) -> write", &client.ToolAnnotations{}, "write"},
		{"readOnly wins over destructive flag", &client.ToolAnnotations{
			ReadOnlyHint: &tr, DestructiveHint: &tr,
		}, "read"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toolCategoryLabel(tt.a); got != tt.want {
				t.Errorf("toolCategoryLabel = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestToolDefinition_AnnotationsRoundTrip(t *testing.T) {
	// Pretend we got this from a tools/list response on the wire and ensure
	// it survives a marshal/unmarshal cycle — the cache writes back via
	// json.Marshal, so we need annotations to persist.
	raw := []byte(`{
		"name": "get_snap_campaigns",
		"description": "Read campaigns",
		"inputSchema": {"type":"object","properties":{}},
		"annotations": {"readOnlyHint": true, "openWorldHint": false}
	}`)

	var td client.ToolDefinition
	if err := json.Unmarshal(raw, &td); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if td.Annotations == nil {
		t.Fatalf("annotations were dropped")
	}
	if td.Annotations.ReadOnlyHint == nil || *td.Annotations.ReadOnlyHint != true {
		t.Errorf("ReadOnlyHint = %v, want *true", td.Annotations.ReadOnlyHint)
	}
	if td.Annotations.OpenWorldHint == nil || *td.Annotations.OpenWorldHint != false {
		t.Errorf("OpenWorldHint = %v, want *false", td.Annotations.OpenWorldHint)
	}

	// Marshal back and verify pointers serialize as bare booleans, not the
	// stringly-typed shape Go can fall into with custom MarshalJSON.
	out, err := json.Marshal(td)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !json.Valid(out) {
		t.Fatalf("re-marshaled JSON is invalid: %s", out)
	}
	var back map[string]interface{}
	if err := json.Unmarshal(out, &back); err != nil {
		t.Fatalf("re-unmarshal: %v", err)
	}
	ann, ok := back["annotations"].(map[string]interface{})
	if !ok {
		t.Fatalf("annotations missing after round trip: %s", out)
	}
	if ann["readOnlyHint"] != true {
		t.Errorf("round-tripped readOnlyHint = %v, want true", ann["readOnlyHint"])
	}
	if ann["openWorldHint"] != false {
		t.Errorf("round-tripped openWorldHint = %v, want false", ann["openWorldHint"])
	}
	// Fields we did not set must not appear at all (omitempty respected).
	if _, present := ann["destructiveHint"]; present {
		t.Errorf("destructiveHint should be omitted when unset")
	}
}

func TestEmitJSONError_InJSONMode(t *testing.T) {
	prev := mcpJSON
	mcpJSON = true
	defer func() { mcpJSON = prev }()

	var retErr error
	out := captureStdout(t, func() {
		retErr = emitJSONError(io.EOF)
	})
	if retErr != io.EOF {
		t.Errorf("expected io.EOF to propagate for exit code, got %v", retErr)
	}

	var got map[string]interface{}
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("stdout not valid JSON: %v\n%s", err, out)
	}
	if got["ok"] != false {
		t.Errorf("ok = %v, want false", got["ok"])
	}
	if got["error"] != "EOF" {
		t.Errorf("error = %v, want EOF", got["error"])
	}
}

func TestToolDefinition_TitleRoundTrip(t *testing.T) {
	// A tools/list response as Pipeboard servers actually send it: `title` is a
	// TOP-LEVEL field (MCP spec 2025-06-18), not `annotations.title`. Decoding
	// used to drop it, so `tools-list` showed every tool as untitled and the
	// on-disk cache persisted that — hence the round-trip assertion.
	raw := []byte(`{
		"name": "execute_google_ads_gaql_query",
		"title": "Execute Google Ads GAQL Query",
		"description": "Runs a GAQL query",
		"inputSchema": {"type":"object","properties":{}},
		"annotations": {"readOnlyHint": true}
	}`)

	var td client.ToolDefinition
	if err := json.Unmarshal(raw, &td); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if td.Title != "Execute Google Ads GAQL Query" {
		t.Fatalf("Title = %q, want the server's top-level title", td.Title)
	}

	out, err := json.Marshal(td)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var back map[string]interface{}
	if err := json.Unmarshal(out, &back); err != nil {
		t.Fatalf("re-unmarshal: %v", err)
	}
	if back["title"] != "Execute Google Ads GAQL Query" {
		t.Errorf("re-marshaled title = %v, want it preserved for --json and the cache", back["title"])
	}
}

func TestToolDisplayTitle(t *testing.T) {
	tests := []struct {
		name string
		td   client.ToolDefinition
		want string
	}{
		{"top-level title", client.ToolDefinition{Title: "Get Campaigns"}, "Get Campaigns"},
		{"legacy annotations.title fallback", client.ToolDefinition{
			Annotations: &client.ToolAnnotations{Title: "Get Campaigns"},
		}, "Get Campaigns"},
		{"top-level wins over annotations", client.ToolDefinition{
			Title:       "Get Campaigns",
			Annotations: &client.ToolAnnotations{Title: "Stale Legacy Title"},
		}, "Get Campaigns"},
		{"neither -> empty, never the machine name", client.ToolDefinition{Name: "get_campaigns"}, ""},
		{"nil annotations", client.ToolDefinition{Name: "get_campaigns", Annotations: nil}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.td.DisplayTitle(); got != tt.want {
				t.Errorf("DisplayTitle() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTitleOrMissing(t *testing.T) {
	if got := titleOrMissing(client.ToolDefinition{Title: "Get Campaigns"}); got != "Get Campaigns" {
		t.Errorf("titleOrMissing = %q, want the title", got)
	}
	// An untitled tool is a blocking submission defect — say so, don't render blank.
	if got := titleOrMissing(client.ToolDefinition{Name: "get_campaigns"}); got != "(no title!)" {
		t.Errorf("titleOrMissing = %q, want the missing-title placeholder", got)
	}
}
