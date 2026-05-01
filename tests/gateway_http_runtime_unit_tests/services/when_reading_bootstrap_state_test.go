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

type WhenReadingBootstrapStateTestingSuite struct {
	sut    *services.HttpBootstrapStateReader
	server *httptest.Server
}

func WhenReadingBootstrapStateBeforeEach(t *testing.T, handler http.HandlerFunc) *WhenReadingBootstrapStateTestingSuite {
	t.Helper()
	server := httptest.NewServer(handler)
	client := clients.NewRuntimeHttpClient(&settings.RuntimeClientSettings{
		BaseUrl: server.URL,
		Timeout: 2 * time.Second,
	})
	return &WhenReadingBootstrapStateTestingSuite{
		sut:    services.NewHttpBootstrapStateReader(client),
		server: server,
	}
}

func WhenReadingBootstrapStateAfterEach(suite *WhenReadingBootstrapStateTestingSuite) {
	if suite.server != nil {
		suite.server.Close()
	}
}

func WhenReadingBootstrapStateBeforeEachUnreachable() *WhenReadingBootstrapStateTestingSuite {
	client := clients.NewRuntimeHttpClient(&settings.RuntimeClientSettings{
		BaseUrl: "http://127.0.0.1:1",
		Timeout: 100 * time.Millisecond,
	})
	return &WhenReadingBootstrapStateTestingSuite{
		sut: services.NewHttpBootstrapStateReader(client),
	}
}

func TestWhenReadingBootstrapState(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime returns a ready state", func(t *testing.T) {
		t.Parallel()

		var seenPath string
		handler := func(w http.ResponseWriter, r *http.Request) {
			seenPath = r.URL.Path
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"status":"ready","agentAddress":"0xabc"}`))
		}

		t.Run("Should hit /internal/bootstrap/status and map the payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingBootstrapStateBeforeEach(t, handler)
			defer WhenReadingBootstrapStateAfterEach(suite)

			state := suite.sut.Get()

			assert.Equal(t, "/internal/bootstrap/status", seenPath)
			assert.Equal(t, enums.BootstrapStatusReady, state.Status)
			assert.Equal(t, "0xabc", state.AgentAddress)
			assert.Empty(t, state.LastError)
		})
	})

	t.Run("Given the runtime is unreachable", func(t *testing.T) {
		t.Parallel()

		t.Run("Should surface the transport error in LastError", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingBootstrapStateBeforeEachUnreachable()
			defer WhenReadingBootstrapStateAfterEach(suite)

			state := suite.sut.Get()

			assert.NotEmpty(t, state.LastError)
		})
	})
}
