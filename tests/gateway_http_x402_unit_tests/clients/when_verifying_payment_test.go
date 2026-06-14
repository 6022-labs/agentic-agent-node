package clients_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/models/x402"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_x402/clients"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type WhenVerifyingPaymentTestingSuite struct {
	sut    *clients.FacilitatorClient
	server *httptest.Server
}

func WhenVerifyingPaymentBeforeEach(t *testing.T, handler http.HandlerFunc) *WhenVerifyingPaymentTestingSuite {
	t.Helper()

	server := httptest.NewServer(handler)
	sut := clients.NewFacilitatorClient(clients.NewFacilitatorClientParams{
		Logger:     zap.NewNop(),
		HTTPClient: http.DefaultClient,
	})

	return &WhenVerifyingPaymentTestingSuite{sut: sut, server: server}
}

func WhenVerifyingPaymentAfterEach(suite *WhenVerifyingPaymentTestingSuite) {
	suite.server.Close()
}

func facilitatorTestPayload() (*x402.PaymentPayload, *x402.PaymentRequirements) {
	payload := &x402.PaymentPayload{
		X402Version: x402.Version,
		Scheme:      x402.SchemeExact,
		Network:     "eip155:80002",
		Payload: x402.ExactEvmPayload{
			Signature: "0xsignature",
			Authorization: &x402.Eip3009Authorization{
				From:        "0x70997970C51812dc3A010C7d01b50e0d17dc79C8",
				To:          "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
				Value:       "10000",
				ValidAfter:  "0",
				ValidBefore: "9999999999",
				Nonce:       "0x0000000000000000000000000000000000000000000000000000000000000001",
			},
		},
	}
	requirements := &x402.PaymentRequirements{
		Scheme:            x402.SchemeExact,
		Network:           "eip155:80002",
		Amount:            "10000",
		Asset:             "0x41E94Eb019C0762f9Bfcf9Fb1E58725BfB0e7582",
		PayTo:             "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
		MaxTimeoutSeconds: 60,
	}
	return payload, requirements
}

func TestWhenVerifyingPayment(t *testing.T) {
	t.Parallel()

	t.Run("Given the facilitator validates the payment", func(t *testing.T) {
		t.Parallel()

		var seenMethod, seenPath, seenContentType string
		var seenBody []byte
		handler := func(w http.ResponseWriter, r *http.Request) {
			seenMethod = r.Method
			seenPath = r.URL.Path
			seenContentType = r.Header.Get("Content-Type")
			seenBody, _ = io.ReadAll(r.Body)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"isValid":true,"payer":"0x70997970C51812dc3A010C7d01b50e0d17dc79C8"}`))
		}

		t.Run("Should POST the v2 envelope to /verify and decode the result", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingPaymentBeforeEach(t, handler)
			defer WhenVerifyingPaymentAfterEach(suite)

			payload, requirements := facilitatorTestPayload()
			resp, err := suite.sut.Verify(context.Background(), suite.server.URL, payload, requirements)

			assert.NoError(t, err)
			if assert.NotNil(t, resp) {
				assert.True(t, resp.IsValid)
				assert.Equal(t, "0x70997970C51812dc3A010C7d01b50e0d17dc79C8", resp.Payer)
			}
			assert.Equal(t, http.MethodPost, seenMethod)
			assert.Equal(t, "/verify", seenPath)
			assert.Equal(t, "application/json", seenContentType)

			var envelope map[string]json.RawMessage
			assert.NoError(t, json.Unmarshal(seenBody, &envelope))
			assert.Equal(t, "2", string(envelope["x402Version"]))
			assert.Contains(t, string(envelope["paymentPayload"]), "eip155:80002")
			assert.Contains(t, string(envelope["paymentRequirements"]), "10000")
		})
	})

	t.Run("Given the facilitator rejects the payment", func(t *testing.T) {
		t.Parallel()

		handler := func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"isValid":false,"invalidReason":"insufficient_funds"}`))
		}

		t.Run("Should return the invalid reason without error", func(t *testing.T) {
			t.Parallel()

			suite := WhenVerifyingPaymentBeforeEach(t, handler)
			defer WhenVerifyingPaymentAfterEach(suite)

			payload, requirements := facilitatorTestPayload()
			resp, err := suite.sut.Verify(context.Background(), suite.server.URL, payload, requirements)

			assert.NoError(t, err)
			assert.False(t, resp.IsValid)
			assert.Equal(t, "insufficient_funds", resp.InvalidReason)
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

			suite := WhenVerifyingPaymentBeforeEach(t, handler)
			defer WhenVerifyingPaymentAfterEach(suite)

			payload, requirements := facilitatorTestPayload()
			resp, err := suite.sut.Verify(context.Background(), suite.server.URL, payload, requirements)

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

			suite := WhenVerifyingPaymentBeforeEach(t, handler)
			defer WhenVerifyingPaymentAfterEach(suite)

			payload, requirements := facilitatorTestPayload()
			resp, err := suite.sut.Verify(context.Background(), suite.server.URL, payload, requirements)

			assert.Nil(t, resp)
			assert.Error(t, err)
		})
	})
}
