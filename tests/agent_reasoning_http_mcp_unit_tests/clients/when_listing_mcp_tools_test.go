package clients_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_mcp/clients"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type WhenListingMcpToolsTestingSuite struct {
	sut    *clients.McpClient
	server *httptest.Server
}

func WhenListingMcpToolsBeforeEach(t *testing.T, handler http.HandlerFunc) *WhenListingMcpToolsTestingSuite {
	t.Helper()
	server := httptest.NewServer(handler)
	sut := clients.NewMcpClient(clients.NewMcpClientParams{
		Logger:     zap.NewNop(),
		HttpClient: http.DefaultClient,
	})
	return &WhenListingMcpToolsTestingSuite{
		sut:    sut,
		server: server,
	}
}

func WhenListingMcpToolsAfterEach(suite *WhenListingMcpToolsTestingSuite) {
	suite.server.Close()
}

func TestWhenListingMcpTools(t *testing.T) {
	t.Parallel()

	t.Run("Given the MCP server returns a tools list", func(t *testing.T) {
		t.Parallel()

		var seenContentType, seenAccept, seenAuth string
		handler := func(w http.ResponseWriter, r *http.Request) {
			seenContentType = r.Header.Get("Content-Type")
			seenAccept = r.Header.Get("Accept")
			seenAuth = r.Header.Get("Authorization")
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"tools":[{"name":"echo","description":"Echo input","inputSchema":{"type":"object","properties":{},"required":[]}}]}}`))
		}

		t.Run("Should POST jsonrpc list tools and forward custom headers", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingMcpToolsBeforeEach(t, handler)
			defer WhenListingMcpToolsAfterEach(suite)

			res, err := suite.sut.ListTools(context.Background(), suite.server.URL, map[string]string{"Authorization": "Bearer token"})

			assert.NoError(t, err)
			assert.Equal(t, "application/json", seenContentType)
			assert.Contains(t, seenAccept, "application/json")
			assert.Equal(t, "Bearer token", seenAuth)
			assert.Len(t, res.Result.Tools, 1)
			assert.Equal(t, "echo", res.Result.Tools[0].Name)
		})
	})

	t.Run("Given the MCP server returns a non-2xx", func(t *testing.T) {
		t.Parallel()

		handler := func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}

		t.Run("Should return an error", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingMcpToolsBeforeEach(t, handler)
			defer WhenListingMcpToolsAfterEach(suite)

			_, err := suite.sut.ListTools(context.Background(), suite.server.URL, nil)
			assert.Error(t, err)
		})
	})

	t.Run("Given the MCP server refuses the auth headers", func(t *testing.T) {
		t.Parallel()

		handler := func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}

		t.Run("Should return the status so the requester can map it", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingMcpToolsBeforeEach(t, handler)
			defer WhenListingMcpToolsAfterEach(suite)

			_, err := suite.sut.ListTools(context.Background(), suite.server.URL, map[string]string{"Authorization": "Bearer wrong"})

			var statusErr *clients.McpStatusError
			assert.ErrorAs(t, err, &statusErr)
			assert.Equal(t, http.StatusUnauthorized, statusErr.StatusCode)
		})
	})

	t.Run("Given the context is already cancelled", func(t *testing.T) {
		t.Parallel()

		handler := func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"tools":[]}}`))
		}

		t.Run("Should abort the request instead of hitting the server", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingMcpToolsBeforeEach(t, handler)
			defer WhenListingMcpToolsAfterEach(suite)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			_, err := suite.sut.ListTools(ctx, suite.server.URL, nil)
			assert.ErrorIs(t, err, context.Canceled)
		})
	})

	t.Run("Given an SSE stream with a trailing keep-alive frame", func(t *testing.T) {
		t.Parallel()

		handler := func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			_, _ = w.Write([]byte("data: {\"jsonrpc\":\"2.0\",\"id\":1,\"result\":{\"tools\":[{\"name\":\"echo\",\"description\":\"Echo input\",\"inputSchema\":{\"type\":\"object\",\"properties\":{},\"required\":[]}}]}}\n\n"))
			_, _ = w.Write([]byte("data: {}\n\n"))
		}

		t.Run("Should return the tools, not the empty trailing frame", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingMcpToolsBeforeEach(t, handler)
			defer WhenListingMcpToolsAfterEach(suite)

			res, err := suite.sut.ListTools(context.Background(), suite.server.URL, nil)

			assert.NoError(t, err)
			assert.Len(t, res.Result.Tools, 1)
			assert.Equal(t, "echo", res.Result.Tools[0].Name)
		})
	})

	t.Run("Given an id-less frame arrives before the id-bearing result frame", func(t *testing.T) {
		t.Parallel()

		handler := func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			_, _ = w.Write([]byte("data: {}\n\n"))
			_, _ = w.Write([]byte("data: {\"jsonrpc\":\"2.0\",\"id\":1,\"result\":{\"tools\":[{\"name\":\"echo\",\"description\":\"Echo input\",\"inputSchema\":{\"type\":\"object\",\"properties\":{},\"required\":[]}}]}}\n\n"))
		}

		t.Run("Should let the real result win over the earlier keep-alive", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingMcpToolsBeforeEach(t, handler)
			defer WhenListingMcpToolsAfterEach(suite)

			res, err := suite.sut.ListTools(context.Background(), suite.server.URL, nil)

			assert.NoError(t, err)
			assert.Len(t, res.Result.Tools, 1)
			assert.Equal(t, "echo", res.Result.Tools[0].Name)
		})
	})

	t.Run("Given only id-less keep-alive frames arrive", func(t *testing.T) {
		t.Parallel()

		handler := func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			_, _ = w.Write([]byte("data: {}\n\n"))
			_, _ = w.Write([]byte("data: {}\n\n"))
		}

		t.Run("Should still return a non-nil frame without crashing", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingMcpToolsBeforeEach(t, handler)
			defer WhenListingMcpToolsAfterEach(suite)

			res, err := suite.sut.ListTools(context.Background(), suite.server.URL, nil)

			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.Empty(t, res.Result.Tools)
		})
	})

	t.Run("Given the MCP server returns a JSON-RPC error frame", func(t *testing.T) {
		t.Parallel()

		handler := func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"error":{"code":-32601,"message":"Method not found"}}`))
		}

		t.Run("Should surface the error payload without returning a transport error", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingMcpToolsBeforeEach(t, handler)
			defer WhenListingMcpToolsAfterEach(suite)

			res, err := suite.sut.ListTools(context.Background(), suite.server.URL, nil)

			assert.NoError(t, err)
			if assert.NotNil(t, res.Error) {
				assert.Equal(t, -32601, res.Error.Code)
				assert.Equal(t, "Method not found", res.Error.Message)
			}
		})
	})
}
