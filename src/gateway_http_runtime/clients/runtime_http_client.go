package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/settings"
)

// RuntimeHttpClient is the low-level HTTP wrapper around runtime_host. Service
// adapters in the same package compose it.
type RuntimeHttpClient struct {
	httpClient *http.Client
	baseUrl    string
}

func NewRuntimeHttpClient(s *settings.RuntimeClientSettings) *RuntimeHttpClient {
	return &RuntimeHttpClient{
		httpClient: &http.Client{Timeout: s.Timeout},
		baseUrl:    strings.TrimRight(s.BaseUrl, "/"),
	}
}

// Get sends a GET and decodes a JSON response into out. notFoundErr is
// returned on 404 if non-nil.
func (c *RuntimeHttpClient) Get(ctx context.Context, path string, out any, notFoundErr error) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseUrl+path, nil)
	if err != nil {
		return err
	}
	return c.do(req, out, notFoundErr)
}

// Post sends a POST with a JSON body and decodes a JSON response into out.
func (c *RuntimeHttpClient) Post(ctx context.Context, path string, body any, out any) error {
	encoded, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseUrl+path, bytes.NewReader(encoded))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req, out, nil)
}

func (c *RuntimeHttpClient) do(req *http.Request, out any, notFoundErr error) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("runtime request failed: %w", err)
	}
	defer resp.Body.Close()

	if notFoundErr != nil && resp.StatusCode == http.StatusNotFound {
		return notFoundErr
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("runtime returned %d: %s", resp.StatusCode, string(body))
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
