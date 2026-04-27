package requests_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/requests"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingAuthorizeRequest(t *testing.T) {
	t.Parallel()

	t.Run("Given a valid 0x-prefixed 20-byte address", func(t *testing.T) {
		t.Parallel()
		req := &requests.AuthorizeRequest{Address: "0x1111111111111111111111111111111111111111"}

		t.Run("Should pass validation", func(t *testing.T) {
			err := req.ValidateAndSanitize()
			assert.NoError(t, err)
		})
	})

	t.Run("Given an address with surrounding whitespace", func(t *testing.T) {
		t.Parallel()
		req := &requests.AuthorizeRequest{Address: "  0xAbCdEf0123456789abcdef0123456789ABCDEF01  "}

		t.Run("Should trim and accept it", func(t *testing.T) {
			err := req.ValidateAndSanitize()
			assert.NoError(t, err)
			assert.Equal(t, "0xAbCdEf0123456789abcdef0123456789ABCDEF01", req.Address)
		})
	})

	t.Run("Given an empty address", func(t *testing.T) {
		t.Parallel()
		req := &requests.AuthorizeRequest{Address: ""}

		t.Run("Should reject with a validation error on the address field", func(t *testing.T) {
			err := req.ValidateAndSanitize()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "address")
		})
	})

	t.Run("Given an address missing the 0x prefix", func(t *testing.T) {
		t.Parallel()
		req := &requests.AuthorizeRequest{Address: "1111111111111111111111111111111111111111"}

		t.Run("Should reject", func(t *testing.T) {
			err := req.ValidateAndSanitize()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "address")
		})
	})

	t.Run("Given an address shorter than 20 bytes", func(t *testing.T) {
		t.Parallel()
		req := &requests.AuthorizeRequest{Address: "0x1234"}

		t.Run("Should reject", func(t *testing.T) {
			err := req.ValidateAndSanitize()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "address")
		})
	})

	t.Run("Given an address with non-hex characters", func(t *testing.T) {
		t.Parallel()
		req := &requests.AuthorizeRequest{Address: "0xZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZ"}

		t.Run("Should reject", func(t *testing.T) {
			err := req.ValidateAndSanitize()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "address")
		})
	})
}
