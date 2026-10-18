package clients_test

import (
	"context"
	"encoding/json"
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

	t.Run("Given the facilitator call fails", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a transport error", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettlingPaymentBeforeEach(t, func(http.ResponseWriter, *http.Request) {})
			WhenSettlingPaymentAfterEach(suite)

			payload, requirements := facilitatorTestPayload()
			resp, err := suite.sut.Settle(context.Background(), suite.server.URL, payload, requirements)

			assert.Nil(t, resp)
			assert.ErrorContains(t, err, "failed to send HTTP request")
		})
	})

	t.Run("Given the facilitator returns a non-2xx status", func(t *testing.T) {
		t.Parallel()

		handler := func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`boom`))
		}

		t.Run("Should return an error carrying the body", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettlingPaymentBeforeEach(t, handler)
			defer WhenSettlingPaymentAfterEach(suite)

			payload, requirements := facilitatorTestPayload()
			resp, err := suite.sut.Settle(context.Background(), suite.server.URL, payload, requirements)

			assert.Nil(t, resp)
			assert.ErrorContains(t, err, "boom")
		})
	})

	t.Run("Given the facilitator returns malformed JSON", func(t *testing.T) {
		t.Parallel()

		handler := func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{not-json`))
		}

		t.Run("Should return a decode error", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettlingPaymentBeforeEach(t, handler)
			defer WhenSettlingPaymentAfterEach(suite)

			payload, requirements := facilitatorTestPayload()
			resp, err := suite.sut.Settle(context.Background(), suite.server.URL, payload, requirements)

			assert.Nil(t, resp)
			assert.ErrorContains(t, err, "failed to decode facilitator response")
		})
	})

	t.Run("Given the facilitator settles the payment", func(t *testing.T) {
		t.Parallel()

		type seenRequest struct {
			method string
			path   string
			body   []byte
		}

		// serve returns a handler that records the request for the Should to assert on.
		serve := func(seen *seenRequest) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				seen.method = r.Method
				seen.path = r.URL.Path
				seen.body, _ = io.ReadAll(r.Body)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write(
					[]byte(
						`{"success":true,"transaction":"0xtxhash","network":"eip155:80002","payer":"0x70997970C51812dc3A010C7d01b50e0d17dc79C8"}`,
					),
				)
			}
		}

		t.Run("Should POST to /settle", func(t *testing.T) {
			t.Parallel()

			seen := &seenRequest{}
			suite := WhenSettlingPaymentBeforeEach(t, serve(seen))
			defer WhenSettlingPaymentAfterEach(suite)

			payload, requirements := facilitatorTestPayload()
			_, err := suite.sut.Settle(context.Background(), suite.server.URL, payload, requirements)

			assert.NoError(t, err)
			assert.Equal(t, http.MethodPost, seen.method)
			assert.Equal(t, "/settle", seen.path)
		})

		t.Run("Should send the v2 envelope", func(t *testing.T) {
			t.Parallel()

			seen := &seenRequest{}
			suite := WhenSettlingPaymentBeforeEach(t, serve(seen))
			defer WhenSettlingPaymentAfterEach(suite)

			payload, requirements := facilitatorTestPayload()
			_, err := suite.sut.Settle(context.Background(), suite.server.URL, payload, requirements)

			assert.NoError(t, err)

			var envelope map[string]json.RawMessage
			assert.NoError(t, json.Unmarshal(seen.body, &envelope))
			assert.Equal(t, "2", string(envelope["x402Version"]))
			assert.Contains(t, string(envelope["paymentPayload"]), "eip155:80002")
			assert.Contains(t, string(envelope["paymentRequirements"]), "10000")
		})

		t.Run("Should decode the transaction", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettlingPaymentBeforeEach(t, serve(&seenRequest{}))
			defer WhenSettlingPaymentAfterEach(suite)

			payload, requirements := facilitatorTestPayload()
			resp, err := suite.sut.Settle(context.Background(), suite.server.URL+"/", payload, requirements)

			assert.NoError(t, err)
			if assert.NotNil(t, resp) {
				assert.True(t, resp.Success)
				assert.Equal(t, "0xtxhash", resp.Transaction)
			}
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
}
