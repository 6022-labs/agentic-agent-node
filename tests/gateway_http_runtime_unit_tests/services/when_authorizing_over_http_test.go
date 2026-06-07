package services_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

type authorizeCaptured struct {
	method string
	path   string
	body   string
}

type WhenAuthorizingOverHttpTestingSuite struct {
	sut      *services.HttpAuthorizer
	server   *httptest.Server
	captured *authorizeCaptured
}

func WhenAuthorizingOverHttpBeforeEach(t *testing.T, handler http.HandlerFunc) *WhenAuthorizingOverHttpTestingSuite {
	t.Helper()
	captured := &authorizeCaptured{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured.method = r.Method
		captured.path = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		captured.body = string(body)
		handler(w, r)
	}))
	client, _ := clients.NewClientWithResponses(server.URL)
	return &WhenAuthorizingOverHttpTestingSuite{
		sut:      services.NewHttpAuthorizer(client),
		server:   server,
		captured: captured,
	}
}

func WhenAuthorizingOverHttpAfterEach(suite *WhenAuthorizingOverHttpTestingSuite) {
	suite.server.Close()
}

func TestWhenAuthorizingOverHttp(t *testing.T) {
	t.Parallel()

	address := common.HexToAddress("0x1111111111111111111111111111111111111111")

	t.Run("Given the runtime authorizes the address", func(t *testing.T) {
		t.Parallel()

		handler := func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"authorized":true}`))
		}

		t.Run("Should POST the address to /internal/authorize and return true", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingOverHttpBeforeEach(t, handler)
			defer WhenAuthorizingOverHttpAfterEach(suite)

			ok, err := suite.sut.IsAuthorized(context.Background(), address)

			assert.NoError(t, err)
			assert.True(t, ok)
			assert.Equal(t, http.MethodPost, suite.captured.method)
			assert.Equal(t, "/internal/authorize", suite.captured.path)

			var sent models.ExternalAuthorizePayload
			_ = json.Unmarshal([]byte(suite.captured.body), &sent)
			assert.Equal(t, address.Hex(), sent.Address)
		})
	})

	t.Run("Given the runtime denies the address", func(t *testing.T) {
		t.Parallel()

		handler := func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
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

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		handler := func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"code":"boom","message":"runtime down"}`))
		}

		t.Run("Should return the error and false", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingOverHttpBeforeEach(t, handler)
			defer WhenAuthorizingOverHttpAfterEach(suite)

			ok, err := suite.sut.IsAuthorized(context.Background(), address)

			assert.Error(t, err)
			assert.False(t, ok)
		})
	})
}
