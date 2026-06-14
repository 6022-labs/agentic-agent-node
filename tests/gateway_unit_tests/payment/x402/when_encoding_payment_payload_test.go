package x402_test

import (
	"encoding/base64"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/x402"
	"github.com/stretchr/testify/assert"
)

// encodedPayloadJSON returns the JSON the payload serializes to on the wire.
func encodedPayloadJSON(t *testing.T, payload *x402.PaymentPayload) string {
	t.Helper()
	header, err := x402.EncodeBase64Json(payload)
	assert.NoError(t, err)
	decoded, err := base64.StdEncoding.DecodeString(header)
	assert.NoError(t, err)
	return string(decoded)
}

func TestWhenEncodingPaymentPayload(t *testing.T) {
	t.Parallel()

	t.Run("Given an eip3009 payment", func(t *testing.T) {
		t.Parallel()

		t.Run("Should carry only the authorization, not permit2Authorization", func(t *testing.T) {
			t.Parallel()

			json := encodedPayloadJSON(t, &x402.PaymentPayload{
				X402Version: x402.Version,
				Scheme:      x402.SchemeExact,
				Network:     "eip155:80002",
				Payload: x402.ExactEvmPayload{
					Signature:     "0xsignature",
					Authorization: &x402.Eip3009Authorization{Value: "10000"},
				},
			})

			assert.Contains(t, json, "authorization")
			assert.NotContains(t, json, "permit2Authorization")
		})
	})

	t.Run("Given a permit2 payment", func(t *testing.T) {
		t.Parallel()

		t.Run("Should carry only permit2Authorization, not the eip3009 authorization", func(t *testing.T) {
			t.Parallel()

			json := encodedPayloadJSON(t, &x402.PaymentPayload{
				X402Version: x402.Version,
				Scheme:      x402.SchemeExact,
				Network:     "eip155:80002",
				Payload: x402.ExactEvmPayload{
					Signature:            "0xsignature",
					Permit2Authorization: &x402.Permit2Authorization{Nonce: "1"},
				},
			})

			assert.Contains(t, json, "permit2Authorization")
			// The empty eip3009 authorization must not leak onto the wire.
			assert.NotContains(t, json, `"authorization"`)
		})
	})
}
