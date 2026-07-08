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

type WhenCallingMcpToolTestingSuite struct {
	sut    *clients.McpClient
	server *httptest.Server
}

func WhenCallingMcpToolBeforeEach(t *testing.T, handler http.HandlerFunc) *WhenCallingMcpToolTestingSuite {
	t.Helper()
	server := httptest.NewServer(handler)
	sut := clients.NewMcpClient(clients.NewMcpClientParams{
		Logger:     zap.NewNop(),
		HttpClient: http.DefaultClient,
	})
	return &WhenCallingMcpToolTestingSuite{
		sut:    sut,
		server: server,
	}
}

func WhenCallingMcpToolAfterEach(suite *WhenCallingMcpToolTestingSuite) {
	suite.server.Close()
}

func TestWhenCallingMcpTool(t *testing.T) {
	t.Parallel()

	t.Run("Given the MCP server returns a tool call result", func(t *testing.T) {
		t.Parallel()

		var seenBody []byte
		handler := func(w http.ResponseWriter, r *http.Request) {
			buf := make([]byte, r.ContentLength)
			_, _ = r.Body.Read(buf)
			seenBody = buf
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":2,"result":{"content":[{"type":"text","text":"echoed"}]}}`))
		}

		t.Run("Should send the tool name and inputs in the jsonrpc params", func(t *testing.T) {
			t.Parallel()

			suite := WhenCallingMcpToolBeforeEach(t, handler)
			defer WhenCallingMcpToolAfterEach(suite)

			res, err := suite.sut.CallTool(context.Background(), suite.server.URL, "echo", map[string]any{"value": "hi"}, nil)

			assert.NoError(t, err)
			assert.Contains(t, string(seenBody), `"name":"echo"`)
			assert.Contains(t, string(seenBody), `"value":"hi"`)
			assert.NotNil(t, res.Result)
		})
	})

	t.Run("Given an SSE stream with a trailing keep-alive frame", func(t *testing.T) {
		t.Parallel()

		handler := func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			_, _ = w.Write([]byte("data: {\"jsonrpc\":\"2.0\",\"id\":1,\"result\":{\"content\":[{\"type\":\"text\",\"text\":\"echoed\"}]}}\n\n"))
			_, _ = w.Write([]byte("data: {}\n\n"))
		}

		t.Run("Should return the real tool result, not the empty trailing frame", func(t *testing.T) {
			t.Parallel()

			suite := WhenCallingMcpToolBeforeEach(t, handler)
			defer WhenCallingMcpToolAfterEach(suite)

			res, err := suite.sut.CallTool(context.Background(), suite.server.URL, "echo", map[string]any{"value": "hi"}, nil)

			assert.NoError(t, err)
			assert.NotEmpty(t, res.Result)
			assert.Contains(t, string(res.Result), "echoed")
		})
	})

	t.Run("Given the MCP server returns a non-2xx", func(t *testing.T) {
		t.Parallel()

		handler := func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}

		t.Run("Should return an error", func(t *testing.T) {
			t.Parallel()

			suite := WhenCallingMcpToolBeforeEach(t, handler)
			defer WhenCallingMcpToolAfterEach(suite)

			_, err := suite.sut.CallTool(context.Background(), suite.server.URL, "echo", nil, nil)
			assert.Error(t, err)
		})
	})

	t.Run("Given the context is already cancelled", func(t *testing.T) {
		t.Parallel()

		handler := func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{}}`))
		}

		t.Run("Should abort the request instead of hitting the server", func(t *testing.T) {
			t.Parallel()

			suite := WhenCallingMcpToolBeforeEach(t, handler)
			defer WhenCallingMcpToolAfterEach(suite)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			_, err := suite.sut.CallTool(ctx, suite.server.URL, "echo", nil, nil)
			assert.ErrorIs(t, err, context.Canceled)
		})
	})
}
