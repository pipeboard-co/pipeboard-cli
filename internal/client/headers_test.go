package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// recordingServer answers every request with a minimal valid body and keeps
// the headers of the most recent request for assertions.
type recordingServer struct {
	*httptest.Server
	last http.Header
}

func newRecordingServer(t *testing.T) *recordingServer {
	t.Helper()
	rs := &recordingServer{}
	rs.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rs.last = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`{"hash":"abc"}`))
			return
		}
		var req struct {
			ID int64 `json:"id"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      req.ID,
			"result":  map[string]interface{}{"tools": []interface{}{}},
		})
	}))
	t.Cleanup(rs.Close)
	return rs
}

func (rs *recordingServer) bypass() (string, bool) {
	v, ok := rs.last[http.CanonicalHeaderKey(bypassHeader)]
	if !ok || len(v) == 0 {
		return "", false
	}
	return v[0], true
}

func TestBypassHeader_SentOnEveryRequestKindWhenConfigured(t *testing.T) {
	t.Setenv(BypassSecretEnv, "s3cret")
	srv := newRecordingServer(t)

	c := New(srv.URL, "tok", "test")
	if _, err := c.ListTools("meta-ads-mcp"); err != nil {
		t.Fatalf("ListTools: %v", err)
	}
	if got, ok := srv.bypass(); !ok || got != "s3cret" {
		t.Errorf("tools/list: bypass header = %q, %v; want s3cret", got, ok)
	}
	if got := srv.last.Get("Authorization"); got != "Bearer tok" {
		t.Errorf("tools/list: Authorization = %q; want Bearer tok", got)
	}
	if got := srv.last.Get("User-Agent"); got != "pipeboard-cli/test" {
		t.Errorf("tools/list: User-Agent = %q", got)
	}

	if err := c.notify("meta-ads-mcp", "notifications/initialized", nil); err != nil {
		t.Fatalf("notify: %v", err)
	}
	if got, ok := srv.bypass(); !ok || got != "s3cret" {
		t.Errorf("notify: bypass header = %q, %v; want s3cret", got, ok)
	}

	rest := NewREST(srv.URL, "tok", "test")
	if _, err := rest.Post("/api/debug/x", map[string]string{"a": "b"}); err != nil {
		t.Fatalf("REST Post: %v", err)
	}
	if got, ok := srv.bypass(); !ok || got != "s3cret" {
		t.Errorf("REST: bypass header = %q, %v; want s3cret", got, ok)
	}

	if _, err := FetchToolsHash(srv.URL, "test"); err != nil {
		t.Fatalf("FetchToolsHash: %v", err)
	}
	if got, ok := srv.bypass(); !ok || got != "s3cret" {
		t.Errorf("tools-hash: bypass header = %q, %v; want s3cret", got, ok)
	}
	if got := srv.last.Get("Authorization"); got != "" {
		t.Errorf("tools-hash must stay unauthenticated, got Authorization=%q", got)
	}
}

func TestBypassHeader_OmittedWhenUnset(t *testing.T) {
	t.Setenv(BypassSecretEnv, "")
	srv := newRecordingServer(t)

	c := New(srv.URL, "tok", "test")
	if _, err := c.ListTools("meta-ads-mcp"); err != nil {
		t.Fatalf("ListTools: %v", err)
	}
	if got, ok := srv.bypass(); ok {
		t.Errorf("bypass header should be absent when env is unset, got %q", got)
	}
	if got := srv.last.Get("Authorization"); got != "Bearer tok" {
		t.Errorf("Authorization = %q; want Bearer tok", got)
	}

	if _, err := FetchToolsHash(srv.URL, "test"); err != nil {
		t.Fatalf("FetchToolsHash: %v", err)
	}
	if got, ok := srv.bypass(); ok {
		t.Errorf("tools-hash: bypass header should be absent, got %q", got)
	}
}
