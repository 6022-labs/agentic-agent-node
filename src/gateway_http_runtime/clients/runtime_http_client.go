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
	return c.send(ctx, http.MethodPost, path, body, out, nil)
}

// Put sends a PUT with a JSON body and decodes a JSON response into out.
// notFoundErr is returned on 404 if non-nil.
func (c *RuntimeHttpClient) Put(ctx context.Context, path string, body any, out any, notFoundErr error) error {
	return c.send(ctx, http.MethodPut, path, body, out, notFoundErr)
}

// Delete sends a DELETE. notFoundErr is returned on 404 if non-nil.
func (c *RuntimeHttpClient) Delete(ctx context.Context, path string, notFoundErr error) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseUrl+path, nil)
	if err != nil {
		return err
	}
	return c.do(req, nil, notFoundErr)
}

func (c *RuntimeHttpClient) send(ctx context.Context, method, path string, body any, out any, notFoundErr error) error {
	encoded, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseUrl+path, bytes.NewReader(encoded))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req, out, notFoundErr)
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
		statusErr := &RuntimeStatusError{StatusCode: resp.StatusCode, Body: string(body)}
		var parsed struct {
			Code    string `json:"code"`
			Field   string `json:"field"`
			Message string `json:"message"`
		}
		if json.Unmarshal(body, &parsed) == nil {
			statusErr.Code = parsed.Code
			statusErr.Field = parsed.Field
			statusErr.Message = parsed.Message
		}
		return statusErr
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// RuntimeStatusError carries a non-2xx runtime response so callers can preserve
// the semantic status (404/409/400). Error() keeps the legacy string form.
type RuntimeStatusError struct {
	StatusCode int
	Code       string
	Field      string
	Message    string
	Body       string
}

func (e *RuntimeStatusError) Error() string {
	return fmt.Sprintf("runtime returned %d: %s", e.StatusCode, e.Body)
}
