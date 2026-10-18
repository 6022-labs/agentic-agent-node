package requests_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingWalletAddressRequest(t *testing.T) {
	t.Parallel()

	const address = "0xAbCdEf0123456789abcdef0123456789ABCDEF01"

	rejectCases := map[string]string{
		"Given an empty address":                   "",
		"Given an address missing the 0x prefix":   "1111111111111111111111111111111111111111",
		"Given an address shorter than 20 bytes":   "0x1234",
		"Given an address with non-hex characters": "0xZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZ",
	}
	for name, value := range rejectCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			req := &requests.WalletAddressRequest{Address: value}

			t.Run("Should reject on the address field", func(t *testing.T) {
				t.Parallel()

				var validationErr *apperrors.ValidationError
				assert.ErrorAs(t, req.ValidateAndSanitize(), &validationErr)
				assert.Equal(t, "address", validationErr.Field)
			})
		})
	}

	t.Run("Given a padded address", func(t *testing.T) {
		t.Parallel()
		req := &requests.WalletAddressRequest{Address: "  " + address + "  "}

		t.Run("Should trim it", func(t *testing.T) {
			t.Parallel()

			assert.NoError(t, req.ValidateAndSanitize())
			assert.Equal(t, address, req.Address)
		})
	})

	t.Run("Given a valid 0x-prefixed 20-byte address", func(t *testing.T) {
		t.Parallel()
		req := &requests.WalletAddressRequest{Address: "0x1111111111111111111111111111111111111111"}

		t.Run("Should pass validation", func(t *testing.T) {
			t.Parallel()

			assert.NoError(t, req.ValidateAndSanitize())
		})
	})
}
