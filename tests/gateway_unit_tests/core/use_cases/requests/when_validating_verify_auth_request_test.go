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

	t.Run("Given address and signature are valid", func(t *testing.T) {
		t.Parallel()
		req := &requests.VerifyAuthRequest{
			Address:   validAddress,
			Signature: validSignature,
		}

		t.Run("Should pass validation", func(t *testing.T) {
			t.Parallel()
			err := req.ValidateAndSanitize()
			assert.NoError(t, err)
		})
	})

	t.Run("Given fields have surrounding whitespace", func(t *testing.T) {
		t.Parallel()
		req := &requests.VerifyAuthRequest{
			Address:   "  " + validAddress + "  ",
			Signature: "  " + validSignature + "  ",
		}

		t.Run("Should trim them all", func(t *testing.T) {
			t.Parallel()
			err := req.ValidateAndSanitize()
			assert.NoError(t, err)
			assert.Equal(t, validAddress, req.Address)
			assert.Equal(t, validSignature, req.Signature)
		})
	})

	t.Run("Given the address is malformed", func(t *testing.T) {
		t.Parallel()
		req := &requests.VerifyAuthRequest{
			Address:   "0xnope",
			Signature: validSignature,
		}

		t.Run("Should reject with a validation error on address", func(t *testing.T) {
			t.Parallel()
			err := req.ValidateAndSanitize()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "address")
		})
	})

	t.Run("Given the signature is shorter than 65 bytes", func(t *testing.T) {
		t.Parallel()
		req := &requests.VerifyAuthRequest{
			Address:   validAddress,
			Signature: "0xdeadbeef",
		}

		t.Run("Should reject with a validation error on signature", func(t *testing.T) {
			t.Parallel()
			err := req.ValidateAndSanitize()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "signature")
		})
	})

	t.Run("Given the signature has non-hex characters", func(t *testing.T) {
		t.Parallel()
		req := &requests.VerifyAuthRequest{
			Address:   validAddress,
			Signature: "0x" + string(make([]byte, 130)),
		}

		t.Run("Should reject", func(t *testing.T) {
			t.Parallel()
			err := req.ValidateAndSanitize()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "signature")
		})
	})

}
