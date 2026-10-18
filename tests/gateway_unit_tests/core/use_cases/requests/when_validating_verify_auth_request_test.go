package requests_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway/core/use_cases/requests"
	"github.com/stretchr/testify/assert"
)

const (
	validAddress   = "0x1111111111111111111111111111111111111111"
	validSignature = "0x" +
		"1111111111111111111111111111111111111111111111111111111111111111" +
		"2222222222222222222222222222222222222222222222222222222222222222" +
		"1b"
)

func TestWhenValidatingVerifyAuthRequest(t *testing.T) {
	t.Parallel()

	generateVerifyAuthRequest := func() *requests.VerifyAuthRequest {
		return &requests.VerifyAuthRequest{
			WalletAddressRequest: requests.WalletAddressRequest{Address: validAddress},
			Signature:            validSignature,
		}
	}

	t.Run("Given a malformed address", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject with a validation error on address", func(t *testing.T) {
			t.Parallel()

			req := generateVerifyAuthRequest()
			req.Address = "0xnope"

			assert.ErrorContains(t, req.ValidateAndSanitize(), "address")
		})
	})

	t.Run("Given the signature is shorter than 65 bytes", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject with a validation error on signature", func(t *testing.T) {
			t.Parallel()

			req := generateVerifyAuthRequest()
			req.Signature = "0xdeadbeef"

			assert.ErrorContains(t, req.ValidateAndSanitize(), "signature")
		})
	})

	t.Run("Given the signature has non-hex characters", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject with a validation error on signature", func(t *testing.T) {
			t.Parallel()

			req := generateVerifyAuthRequest()
			req.Signature = "0x" + string(make([]byte, 130))

			assert.ErrorContains(t, req.ValidateAndSanitize(), "signature")
		})
	})

	t.Run("Given a padded signature", func(t *testing.T) {
		t.Parallel()

		t.Run("Should trim it", func(t *testing.T) {
			t.Parallel()

			req := generateVerifyAuthRequest()
			req.Signature = "  " + validSignature + "  "

			assert.NoError(t, req.ValidateAndSanitize())
			assert.Equal(t, validSignature, req.Signature)
		})
	})

	t.Run("Given a fully populated request", func(t *testing.T) {
		t.Parallel()

		t.Run("Should pass validation", func(t *testing.T) {
			t.Parallel()

			assert.NoError(t, generateVerifyAuthRequest().ValidateAndSanitize())
		})
	})
}
