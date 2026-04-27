package clients_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/settings"
	"github.com/stretchr/testify/assert"
)

type capturedRequest struct {
	method string
	path   string
	body   string
	auth   string
}

func newClient(t *testing.T, handler http.HandlerFunc) (*clients.RuntimeHttpClient, *capturedRequest, func()) {
	t.Helper()
	captured := &capturedRequest{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured.method = r.Method
		captured.path = r.URL.Path
		captured.auth = r.Header.Get("Authorization")
		body, _ := io.ReadAll(r.Body)
		captured.body = string(body)
		handler(w, r)
	}))
	client := clients.NewRuntimeHttpClient(&settings.RuntimeClientSettings{
		BaseUrl: server.URL,
		Timeout: 2 * time.Second,
	})
	return client, captured, server.Close
}

func TestWhenCallingRuntimeHttpClient(t *testing.T) {
	t.Parallel()

	t.Run("Given a 200 response on GET", func(t *testing.T) {
		t.Parallel()
		client, captured, cleanup := newClient(t, func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"value":"ok"}`))
		})
		defer cleanup()

		t.Run("Should send GET to the path and decode the body", func(t *testing.T) {
			var out struct {
				Value string `json:"value"`
			}
			err := client.Get(context.Background(), "/internal/anything", &out, nil)
			assert.NoError(t, err)
			assert.Equal(t, "ok", out.Value)
			assert.Equal(t, http.MethodGet, captured.method)
			assert.Equal(t, "/internal/anything", captured.path)
		})
	})

	t.Run("Given a 404 response and a notFoundErr is supplied", func(t *testing.T) {
		t.Parallel()
		sentinel := errors.New("not here")
		client, _, cleanup := newClient(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})
		defer cleanup()

		t.Run("Should return the sentinel error", func(t *testing.T) {
			err := client.Get(context.Background(), "/internal/missing", nil, sentinel)
			assert.ErrorIs(t, err, sentinel)
		})
	})

	t.Run("Given a 500 response", func(t *testing.T) {
		t.Parallel()
		client, _, cleanup := newClient(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"code":"boom"}`))
		})
		defer cleanup()

		t.Run("Should surface the status + body in the error", func(t *testing.T) {
			err := client.Get(context.Background(), "/internal/x", nil, nil)
			assert.ErrorContains(t, err, "500")
			assert.ErrorContains(t, err, "boom")
		})
	})

	t.Run("Given a POST with a JSON body", func(t *testing.T) {
		t.Parallel()
		client, captured, cleanup := newClient(t, func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"echoed":true}`))
		})
		defer cleanup()

		t.Run("Should send the body and decode the response", func(t *testing.T) {
			var out struct {
				Echoed bool `json:"echoed"`
			}
			err := client.Post(context.Background(), "/internal/post", map[string]string{"k": "v"}, &out)
			assert.NoError(t, err)
			assert.True(t, out.Echoed)
			assert.Equal(t, http.MethodPost, captured.method)
			assert.Equal(t, "/internal/post", captured.path)
			assert.Contains(t, captured.body, `"k":"v"`)
		})
	})
}
