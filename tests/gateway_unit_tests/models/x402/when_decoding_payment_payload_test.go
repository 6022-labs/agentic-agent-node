package x402_test

import (
	"encoding/base64"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway/models/x402"
	"github.com/stretchr/testify/assert"
)

func TestWhenDecodingPaymentPayload(t *testing.T) {
	t.Parallel()

	t.Run("Given a valid v2 payment header", func(t *testing.T) {
		t.Parallel()

		t.Run("Should round-trip through encode and decode", func(t *testing.T) {
			t.Parallel()

			header, err := x402.EncodeBase64Json(&x402.PaymentPayload{
				X402Version: x402.Version,
				Scheme:      x402.SchemeExact,
				Network:     "eip155:80002",
				Payload: x402.ExactEvmPayload{
					Signature: "0xsignature",
					Authorization: &x402.Eip3009Authorization{
						From:  "0x70997970C51812dc3A010C7d01b50e0d17dc79C8",
						To:    "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
						Value: "10000",
					},
				},
			})
			assert.NoError(t, err)

			payload, err := x402.DecodePaymentPayload(header)

			assert.NoError(t, err)
			assert.Equal(t, x402.SchemeExact, payload.Scheme)
			assert.Equal(t, "eip155:80002", payload.Network)
			assert.Equal(t, "10000", payload.Payload.Authorization.Value)
		})
	})

	t.Run("Given the header is not base64", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return an error", func(t *testing.T) {
			t.Parallel()

			payload, err := x402.DecodePaymentPayload("not-base64!!")

			assert.Nil(t, payload)
			assert.Error(t, err)
		})
	})

	t.Run("Given the header is not JSON", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return an error", func(t *testing.T) {
			t.Parallel()

			header := base64.StdEncoding.EncodeToString([]byte("{not-json"))
			payload, err := x402.DecodePaymentPayload(header)

			assert.Nil(t, payload)
			assert.Error(t, err)
		})
	})

	t.Run("Given an unsupported x402 version", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return an error", func(t *testing.T) {
			t.Parallel()

			header, err := x402.EncodeBase64Json(&x402.PaymentPayload{X402Version: 1})
			assert.NoError(t, err)

			payload, err := x402.DecodePaymentPayload(header)

			assert.Nil(t, payload)
			assert.ErrorContains(t, err, "unsupported x402 version")
		})
	})
}

func TestWhenParsingEip155ChainId(t *testing.T) {
	t.Parallel()

	t.Run("Given a CAIP-2 eip155 identifier", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return the chain id", func(t *testing.T) {
			t.Parallel()

			chainId, err := x402.ParseEip155ChainId("eip155:80002")

			assert.NoError(t, err)
			assert.Equal(t, int64(80002), chainId)
		})
	})

	t.Run("Given the eip155 prefix is missing", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return an error", func(t *testing.T) {
			t.Parallel()

			_, err := x402.ParseEip155ChainId("solana:mainnet")

			assert.Error(t, err)
		})
	})

	t.Run("Given the chain id is not a positive integer", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return an error", func(t *testing.T) {
			t.Parallel()

			_, err := x402.ParseEip155ChainId("eip155:abc")
			assert.Error(t, err)

			_, err = x402.ParseEip155ChainId("eip155:-5")
			assert.Error(t, err)
		})
	})
}
