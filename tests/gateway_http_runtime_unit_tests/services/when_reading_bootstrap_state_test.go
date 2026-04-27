package services_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/common/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/settings"
	"github.com/stretchr/testify/assert"
)

func newReader(t *testing.T, handler http.HandlerFunc) (*services.HttpBootstrapStateReader, func()) {
	t.Helper()
	server := httptest.NewServer(handler)
	client := clients.NewRuntimeHttpClient(&settings.RuntimeClientSettings{
		BaseUrl: server.URL,
		Timeout: 2 * time.Second,
	})
	return services.NewHttpBootstrapStateReader(client), server.Close
}

func TestWhenReadingBootstrapState(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime returns a ready state", func(t *testing.T) {
		t.Parallel()
		var seenPath string
		reader, cleanup := newReader(t, func(w http.ResponseWriter, r *http.Request) {
			seenPath = r.URL.Path
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"status":"ready","agentAddress":"0xabc"}`))
		})
		defer cleanup()

		t.Run("Should hit /internal/bootstrap/status and map the payload", func(t *testing.T) {
			state := reader.Get()
			assert.Equal(t, "/internal/bootstrap/status", seenPath)
			assert.Equal(t, enums.BootstrapStatusReady, state.Status)
			assert.Equal(t, "0xabc", state.AgentAddress)
			assert.Empty(t, state.LastError)
		})
	})

	t.Run("Given the runtime is unreachable", func(t *testing.T) {
		t.Parallel()
		client := clients.NewRuntimeHttpClient(&settings.RuntimeClientSettings{
			BaseUrl: "http://127.0.0.1:1",
			Timeout: 100 * time.Millisecond,
		})
		reader := services.NewHttpBootstrapStateReader(client)

		t.Run("Should surface the transport error in LastError", func(t *testing.T) {
			state := reader.Get()
			assert.NotEmpty(t, state.LastError)
		})
	})
}
