package clients_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway_http_x402/clients"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type WhenSettlingPaymentTestingSuite struct {
	sut    *clients.FacilitatorClient
	server *httptest.Server
}

func WhenSettlingPaymentBeforeEach(t *testing.T, handler http.HandlerFunc) *WhenSettlingPaymentTestingSuite {
	t.Helper()

	server := httptest.NewServer(handler)
	sut := clients.NewFacilitatorClient(clients.NewFacilitatorClientParams{
		Logger:     zap.NewNop(),
		HTTPClient: http.DefaultClient,
	})

	return &WhenSettlingPaymentTestingSuite{sut: sut, server: server}
}

func WhenSettlingPaymentAfterEach(suite *WhenSettlingPaymentTestingSuite) {
	suite.server.Close()
}

func TestWhenSettlingPayment(t *testing.T) {
	t.Parallel()

	t.Run("Given the facilitator settles the payment", func(t *testing.T) {
		t.Parallel()

		var seenMethod, seenPath string
		var seenBody []byte
		handler := func(w http.ResponseWriter, r *http.Request) {
			seenMethod = r.Method
			seenPath = r.URL.Path
			seenBody, _ = io.ReadAll(r.Body)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(
				[]byte(
					`{"success":true,"transaction":"0xtxhash","network":"eip155:80002","payer":"0x70997970C51812dc3A010C7d01b50e0d17dc79C8"}`,
				),
			)
		}

		t.Run("Should POST to /settle and decode the transaction", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettlingPaymentBeforeEach(t, handler)
			defer WhenSettlingPaymentAfterEach(suite)

			payload, requirements := facilitatorTestPayload()
			resp, err := suite.sut.Settle(context.Background(), suite.server.URL+"/", payload, requirements)

			assert.NoError(t, err)
			if assert.NotNil(t, resp) {
				assert.True(t, resp.Success)
				assert.Equal(t, "0xtxhash", resp.Transaction)
			}
			assert.Equal(t, http.MethodPost, seenMethod)
			assert.Equal(t, "/settle", seenPath)
			assert.Contains(t, string(seenBody), `"x402Version":2`)
		})
	})

	t.Run("Given the settlement fails on-chain", func(t *testing.T) {
		t.Parallel()

		handler := func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success":false,"errorReason":"insufficient_funds"}`))
		}

		t.Run("Should return the failure without error", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettlingPaymentBeforeEach(t, handler)
			defer WhenSettlingPaymentAfterEach(suite)

			payload, requirements := facilitatorTestPayload()
			resp, err := suite.sut.Settle(context.Background(), suite.server.URL, payload, requirements)

			assert.NoError(t, err)
			assert.False(t, resp.Success)
			assert.Equal(t, "insufficient_funds", resp.ErrorReason)
		})
	})

	t.Run("Given the facilitator is unreachable", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a transport error", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettlingPaymentBeforeEach(t, func(http.ResponseWriter, *http.Request) {})
			WhenSettlingPaymentAfterEach(suite)

			payload, requirements := facilitatorTestPayload()
			resp, err := suite.sut.Settle(context.Background(), suite.server.URL, payload, requirements)

			assert.Nil(t, resp)
			assert.Error(t, err)
		})
	})
}
