package requests_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway/core/use_cases/requests"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingWalletAddressRequest(t *testing.T) {
	t.Parallel()

	t.Run("Given a malformed address", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject on the address field", func(t *testing.T) {
			t.Parallel()

			req := &requests.WalletAddressRequest{Address: "0x1234"}

			assert.ErrorContains(t, req.ValidateAndSanitize(), "address")
		})
	})

	t.Run("Given a padded address", func(t *testing.T) {
		t.Parallel()

		t.Run("Should trim it", func(t *testing.T) {
			t.Parallel()

			req := &requests.WalletAddressRequest{Address: "  " + validAddress + "  "}

			assert.NoError(t, req.ValidateAndSanitize())
			assert.Equal(t, validAddress, req.Address)
		})
	})
}
