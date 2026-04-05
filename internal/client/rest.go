package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// RESTClient calls Pipeboard REST API endpoints.
type RESTClient struct {
	baseURL    string
	token      string
	userAgent  string
	httpClient *http.Client
}

// NewREST creates a REST API client. version is used in the User-Agent header.
func NewREST(baseURL, token, version string) *RESTClient {
	return &RESTClient{
		baseURL:    baseURL,
		token:      token,
		userAgent:  "pipeboard-cli/" + version,
		httpClient: &http.Client{},
	}
}

// Post sends a JSON POST request and returns the parsed response.
func (c *RESTClient) Post(path string, body interface{}) (map[string]interface{}, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	url := fmt.Sprintf("%s%s", c.baseURL, path)
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		desc, _ := result["error_description"].(string)
		if desc == "" {
			desc, _ = result["error"].(string)
		}
		return nil, fmt.Errorf("%s (HTTP %d)", desc, resp.StatusCode)
	}

	return result, nil
}
