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
