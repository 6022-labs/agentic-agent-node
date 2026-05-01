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

type WhenCallingRuntimeHttpClientTestingSuite struct {
	sut      *clients.RuntimeHttpClient
	server   *httptest.Server
	captured *capturedRequest
}

func WhenCallingRuntimeHttpClientBeforeEach(t *testing.T, handler http.HandlerFunc) *WhenCallingRuntimeHttpClientTestingSuite {
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
	sut := clients.NewRuntimeHttpClient(&settings.RuntimeClientSettings{
		BaseUrl: server.URL,
		Timeout: 2 * time.Second,
	})
	return &WhenCallingRuntimeHttpClientTestingSuite{
		sut:      sut,
		server:   server,
		captured: captured,
	}
}

func WhenCallingRuntimeHttpClientAfterEach(suite *WhenCallingRuntimeHttpClientTestingSuite) {
	suite.server.Close()
}

func TestWhenCallingRuntimeHttpClient(t *testing.T) {
	t.Parallel()

	t.Run("Given a 200 response on GET", func(t *testing.T) {
		t.Parallel()

		handler := func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"value":"ok"}`))
		}

		t.Run("Should send GET to the path and decode the body", func(t *testing.T) {
			t.Parallel()

			suite := WhenCallingRuntimeHttpClientBeforeEach(t, handler)
			defer WhenCallingRuntimeHttpClientAfterEach(suite)

			var out struct {
				Value string `json:"value"`
			}
			err := suite.sut.Get(context.Background(), "/internal/anything", &out, nil)

			assert.NoError(t, err)
			assert.Equal(t, "ok", out.Value)
			assert.Equal(t, http.MethodGet, suite.captured.method)
			assert.Equal(t, "/internal/anything", suite.captured.path)
		})
	})

	t.Run("Given a 404 response and a notFoundErr is supplied", func(t *testing.T) {
		t.Parallel()

		sentinel := errors.New("not here")
		handler := func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}

		t.Run("Should return the sentinel error", func(t *testing.T) {
			t.Parallel()

			suite := WhenCallingRuntimeHttpClientBeforeEach(t, handler)
			defer WhenCallingRuntimeHttpClientAfterEach(suite)

			err := suite.sut.Get(context.Background(), "/internal/missing", nil, sentinel)

			assert.ErrorIs(t, err, sentinel)
		})
	})

	t.Run("Given a non-2xx response", func(t *testing.T) {
		t.Parallel()

		handler := func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"code":"boom"}`))
		}

		t.Run("Should surface the status + body in the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenCallingRuntimeHttpClientBeforeEach(t, handler)
			defer WhenCallingRuntimeHttpClientAfterEach(suite)

			err := suite.sut.Get(context.Background(), "/internal/x", nil, nil)

			assert.ErrorContains(t, err, "500")
			assert.ErrorContains(t, err, "boom")
		})
	})

	t.Run("Given a POST with a JSON body", func(t *testing.T) {
		t.Parallel()

		handler := func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"echoed":true}`))
		}

		t.Run("Should send the body and decode the response", func(t *testing.T) {
			t.Parallel()

			suite := WhenCallingRuntimeHttpClientBeforeEach(t, handler)
			defer WhenCallingRuntimeHttpClientAfterEach(suite)

			var out struct {
				Echoed bool `json:"echoed"`
			}
			err := suite.sut.Post(context.Background(), "/internal/post", map[string]string{"k": "v"}, &out)

			assert.NoError(t, err)
			assert.True(t, out.Echoed)
			assert.Equal(t, http.MethodPost, suite.captured.method)
			assert.Equal(t, "/internal/post", suite.captured.path)
			assert.Contains(t, suite.captured.body, `"k":"v"`)
		})
	})
}
