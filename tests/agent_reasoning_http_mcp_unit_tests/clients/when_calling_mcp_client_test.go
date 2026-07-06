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

type WhenCallingMcpClientTestingSuite struct {
	sut    *clients.McpClient
	server *httptest.Server
}

func WhenCallingMcpClientBeforeEach(t *testing.T, handler http.HandlerFunc) *WhenCallingMcpClientTestingSuite {
	t.Helper()
	server := httptest.NewServer(handler)
	sut := clients.NewMcpClient(clients.NewMcpClientParams{
		Logger:     zap.NewNop(),
		HttpClient: http.DefaultClient,
	})
	return &WhenCallingMcpClientTestingSuite{
		sut:    sut,
		server: server,
	}
}

func WhenCallingMcpClientAfterEach(suite *WhenCallingMcpClientTestingSuite) {
	suite.server.Close()
}

func TestWhenCallingMcpClient(t *testing.T) {
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

			suite := WhenCallingMcpClientBeforeEach(t, handler)
			defer WhenCallingMcpClientAfterEach(suite)

			res, err := suite.sut.ListTools(context.Background(), suite.server.URL, map[string]string{"Authorization": "Bearer token"})

			assert.NoError(t, err)
			assert.Equal(t, "application/json", seenContentType)
			assert.Contains(t, seenAccept, "application/json")
			assert.Equal(t, "Bearer token", seenAuth)
			assert.Len(t, res.Result.Tools, 1)
			assert.Equal(t, "echo", res.Result.Tools[0].Name)
		})
	})

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

			suite := WhenCallingMcpClientBeforeEach(t, handler)
			defer WhenCallingMcpClientAfterEach(suite)

			res, err := suite.sut.CallTool(context.Background(), suite.server.URL, "echo", map[string]any{"value": "hi"}, nil)

			assert.NoError(t, err)
			assert.Contains(t, string(seenBody), `"name":"echo"`)
			assert.Contains(t, string(seenBody), `"value":"hi"`)
			assert.NotNil(t, res.Result)
		})
	})

	t.Run("Given the MCP server returns a non-2xx", func(t *testing.T) {
		t.Parallel()

		handler := func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}

		t.Run("Should return an error from ListTools", func(t *testing.T) {
			t.Parallel()

			suite := WhenCallingMcpClientBeforeEach(t, handler)
			defer WhenCallingMcpClientAfterEach(suite)

			_, err := suite.sut.ListTools(context.Background(), suite.server.URL, nil)
			assert.Error(t, err)
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

			suite := WhenCallingMcpClientBeforeEach(t, handler)
			defer WhenCallingMcpClientAfterEach(suite)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			_, err := suite.sut.ListTools(ctx, suite.server.URL, nil)
			assert.ErrorIs(t, err, context.Canceled)
		})
	})
}
