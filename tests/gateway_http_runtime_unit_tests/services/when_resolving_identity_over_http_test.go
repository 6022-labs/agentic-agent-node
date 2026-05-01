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

type WhenResolvingIdentityOverHttpTestingSuite struct {
	sut    *services.HttpAgentIdentityResolver
	server *httptest.Server
}

func WhenResolvingIdentityOverHttpBeforeEach(t *testing.T, handler http.HandlerFunc) *WhenResolvingIdentityOverHttpTestingSuite {
	t.Helper()
	server := httptest.NewServer(handler)
	client := clients.NewRuntimeHttpClient(&settings.RuntimeClientSettings{
		BaseUrl: server.URL,
		Timeout: 2 * time.Second,
	})
	return &WhenResolvingIdentityOverHttpTestingSuite{
		sut:    services.NewHttpAgentIdentityResolver(client),
		server: server,
	}
}

func WhenResolvingIdentityOverHttpAfterEach(suite *WhenResolvingIdentityOverHttpTestingSuite) {
	suite.server.Close()
}

func TestWhenResolvingIdentityOverHttp(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime returns 200 with identity", func(t *testing.T) {
		t.Parallel()

		var seenPath string
		handler := func(w http.ResponseWriter, r *http.Request) {
			seenPath = r.URL.Path
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"collectionAddress":"0xCOL","tokenId":42,"name":"agent.eth"}`))
		}

		t.Run("Should call /internal/identity and map the payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenResolvingIdentityOverHttpBeforeEach(t, handler)
			defer WhenResolvingIdentityOverHttpAfterEach(suite)

			id, err := suite.sut.Resolve(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, "/internal/identity", seenPath)
			assert.Equal(t, "0xCOL", id.CollectionAddress)
			assert.Equal(t, uint64(42), id.TokenId)
			assert.Equal(t, "agent.eth", id.Name)
		})
	})

	t.Run("Given the runtime returns 404", func(t *testing.T) {
		t.Parallel()

		handler := func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}

		t.Run("Should return ErrIdentityNotFound", func(t *testing.T) {
			t.Parallel()

			suite := WhenResolvingIdentityOverHttpBeforeEach(t, handler)
			defer WhenResolvingIdentityOverHttpAfterEach(suite)

			_, err := suite.sut.Resolve(context.Background())

			assert.ErrorIs(t, err, common_interfaces.ErrIdentityNotFound)
		})
	})

	t.Run("Given the runtime returns a non-2xx", func(t *testing.T) {
		t.Parallel()

		handler := func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}

		t.Run("Should propagate the status error", func(t *testing.T) {
			t.Parallel()

			suite := WhenResolvingIdentityOverHttpBeforeEach(t, handler)
			defer WhenResolvingIdentityOverHttpAfterEach(suite)

			_, err := suite.sut.Resolve(context.Background())

			assert.Error(t, err)
		})
	})
}
