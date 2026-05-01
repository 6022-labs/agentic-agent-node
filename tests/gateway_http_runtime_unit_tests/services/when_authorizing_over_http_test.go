package services_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/settings"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

type WhenAuthorizingOverHttpTestingSuite struct {
	sut    *services.HttpAuthorizer
	server *httptest.Server
}

func WhenAuthorizingOverHttpBeforeEach(t *testing.T, handler http.HandlerFunc) *WhenAuthorizingOverHttpTestingSuite {
	t.Helper()
	server := httptest.NewServer(handler)
	client := clients.NewRuntimeHttpClient(&settings.RuntimeClientSettings{
		BaseUrl: server.URL,
		Timeout: 2 * time.Second,
	})
	return &WhenAuthorizingOverHttpTestingSuite{
		sut:    services.NewHttpAuthorizer(client),
		server: server,
	}
}

func WhenAuthorizingOverHttpAfterEach(suite *WhenAuthorizingOverHttpTestingSuite) {
	suite.server.Close()
}

func TestWhenAuthorizingOverHttp(t *testing.T) {
	t.Parallel()

	address := common.HexToAddress("0x1111111111111111111111111111111111111111")

	t.Run("Given the runtime returns authorized=true", func(t *testing.T) {
		t.Parallel()

		var seenPath, seenBody string
		handler := func(w http.ResponseWriter, r *http.Request) {
			seenPath = r.URL.Path
			body, _ := io.ReadAll(r.Body)
			seenBody = string(body)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"authorized":true}`))
		}

		t.Run("Should POST address to /internal/authorize", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingOverHttpBeforeEach(t, handler)
			defer WhenAuthorizingOverHttpAfterEach(suite)

			ok, err := suite.sut.IsAuthorized(context.Background(), address)

			assert.NoError(t, err)
			assert.True(t, ok)
			assert.Equal(t, "/internal/authorize", seenPath)
			assert.Contains(t, seenBody, address.Hex())
		})
	})

	t.Run("Given the runtime returns authorized=false", func(t *testing.T) {
		t.Parallel()

		handler := func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"authorized":false}`))
		}

		t.Run("Should propagate the false answer", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingOverHttpBeforeEach(t, handler)
			defer WhenAuthorizingOverHttpAfterEach(suite)

			ok, err := suite.sut.IsAuthorized(context.Background(), address)

			assert.NoError(t, err)
			assert.False(t, ok)
		})
	})

	t.Run("Given the runtime returns a non-2xx", func(t *testing.T) {
		t.Parallel()

		handler := func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}

		t.Run("Should return a transport error", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingOverHttpBeforeEach(t, handler)
			defer WhenAuthorizingOverHttpAfterEach(suite)

			ok, err := suite.sut.IsAuthorized(context.Background(), address)

			assert.False(t, ok)
			assert.Error(t, err)
		})
	})
}
