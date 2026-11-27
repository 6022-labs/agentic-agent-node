package transports_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_a2a/settings"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_a2a/transports"
	"github.com/stretchr/testify/assert"
)

type WhenRoundTrippingAnA2ACallTestingSuite struct {
	sut *transports.A2ARoundTripper
}

func WhenRoundTrippingAnA2ACallBeforeEach(
	t *testing.T,
	allowPrivateNetworks bool,
) *WhenRoundTrippingAnA2ACallTestingSuite {
	t.Helper()

	return &WhenRoundTrippingAnA2ACallTestingSuite{
		sut: transports.NewA2ARoundTripper(
			&settings.A2ASettings{RequestTimeoutSeconds: 5, AllowPrivateNetworks: allowPrivateNetworks},
		),
	}
}

func (suite *WhenRoundTrippingAnA2ACallTestingSuite) get(t *testing.T, url string) (*http.Response, error) {
	t.Helper()

	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	assert.NoError(t, err)
	response, err := suite.sut.RoundTrip(request)
	if response != nil {
		t.Cleanup(func() { _ = response.Body.Close() })
	}
	return response, err
}

func TestWhenRoundTrippingAnA2ACall(t *testing.T) {
	t.Parallel()

	t.Run("Given the peer speaks plain http", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse to send the request", func(t *testing.T) {
			t.Parallel()

			var reached atomic.Bool
			server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				reached.Store(true)
			}))
			defer server.Close()
			suite := WhenRoundTrippingAnA2ACallBeforeEach(t, false)

			_, err := suite.get(t, server.URL)

			assert.ErrorIs(t, err, transports.ErrA2AInsecureScheme)
			assert.False(t, reached.Load())
		})
	})

	t.Run("Given the peer resolves to a loopback address", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse to dial it", func(t *testing.T) {
			t.Parallel()

			var reached atomic.Bool
			server := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				reached.Store(true)
			}))
			defer server.Close()
			suite := WhenRoundTrippingAnA2ACallBeforeEach(t, false)

			_, err := suite.get(t, server.URL)

			assert.ErrorIs(t, err, transports.ErrA2AAddressRefused)
			assert.False(t, reached.Load())
		})
	})

	t.Run("Given a private address", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse to dial it", func(t *testing.T) {
			t.Parallel()

			suite := WhenRoundTrippingAnA2ACallBeforeEach(t, false)

			_, err := suite.get(t, "https://10.0.0.1/api/a2a")

			assert.ErrorIs(t, err, transports.ErrA2AAddressRefused)
		})
	})

	t.Run("Given a link-local metadata address", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse to dial it", func(t *testing.T) {
			t.Parallel()

			suite := WhenRoundTrippingAnA2ACallBeforeEach(t, false)

			_, err := suite.get(t, "https://169.254.169.254/latest/meta-data")

			assert.ErrorIs(t, err, transports.ErrA2AAddressRefused)
		})
	})

	t.Run("Given a carrier-grade NAT address", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse to dial it", func(t *testing.T) {
			t.Parallel()

			suite := WhenRoundTrippingAnA2ACallBeforeEach(t, false)

			_, err := suite.get(t, "https://100.64.0.1/api/a2a")

			assert.ErrorIs(t, err, transports.ErrA2AAddressRefused)
		})
	})

	t.Run("Given a loopback address mapped into IPv6", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse to dial it", func(t *testing.T) {
			t.Parallel()

			suite := WhenRoundTrippingAnA2ACallBeforeEach(t, false)

			_, err := suite.get(t, "https://[::ffff:127.0.0.1]/api/a2a")

			assert.ErrorIs(t, err, transports.ErrA2AAddressRefused)
		})
	})

	t.Run("Given a loopback address embedded in NAT64", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse to dial it", func(t *testing.T) {
			t.Parallel()

			suite := WhenRoundTrippingAnA2ACallBeforeEach(t, false)

			_, err := suite.get(t, "https://[64:ff9b::7f00:1]/api/a2a")

			assert.ErrorIs(t, err, transports.ErrA2AAddressRefused)
		})
	})

	t.Run("Given private networks are allowed", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reach the loopback peer over plain http", func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			}))
			defer server.Close()
			suite := WhenRoundTrippingAnA2ACallBeforeEach(t, true)

			response, err := suite.get(t, server.URL)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusNoContent, response.StatusCode)
		})
	})
}
