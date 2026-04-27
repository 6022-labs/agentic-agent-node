package services_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	common_interfaces "github.com/6022-labs/agentic-agent-node/src/common/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/settings"
	"github.com/stretchr/testify/assert"
)

func newResolver(t *testing.T, handler http.HandlerFunc) (*services.HttpAgentIdentityResolver, func()) {
	t.Helper()
	server := httptest.NewServer(handler)
	client := clients.NewRuntimeHttpClient(&settings.RuntimeClientSettings{
		BaseUrl: server.URL,
		Timeout: 2 * time.Second,
	})
	return services.NewHttpAgentIdentityResolver(client), server.Close
}

func TestWhenResolvingIdentityOverHttp(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime returns 200 with identity", func(t *testing.T) {
		t.Parallel()
		var seenPath string
		resolver, cleanup := newResolver(t, func(w http.ResponseWriter, r *http.Request) {
			seenPath = r.URL.Path
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"collectionAddress":"0xCOL","tokenId":42,"name":"agent.eth"}`))
		})
		defer cleanup()

		t.Run("Should call /internal/identity and map the payload", func(t *testing.T) {
			id, err := resolver.Resolve(context.Background())
			assert.NoError(t, err)
			assert.Equal(t, "/internal/identity", seenPath)
			assert.Equal(t, "0xCOL", id.CollectionAddress)
			assert.Equal(t, uint64(42), id.TokenId)
			assert.Equal(t, "agent.eth", id.Name)
		})
	})

	t.Run("Given the runtime returns 404", func(t *testing.T) {
		t.Parallel()
		resolver, cleanup := newResolver(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})
		defer cleanup()

		t.Run("Should return ErrIdentityNotFound", func(t *testing.T) {
			_, err := resolver.Resolve(context.Background())
			assert.ErrorIs(t, err, common_interfaces.ErrIdentityNotFound)
		})
	})

	t.Run("Given the runtime returns 500", func(t *testing.T) {
		t.Parallel()
		resolver, cleanup := newResolver(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})
		defer cleanup()

		t.Run("Should propagate the status error", func(t *testing.T) {
			_, err := resolver.Resolve(context.Background())
			assert.Error(t, err)
		})
	})
}
