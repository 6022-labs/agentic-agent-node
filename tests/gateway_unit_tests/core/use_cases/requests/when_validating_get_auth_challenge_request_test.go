package requests_test

import (
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/use_cases/requests"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingGetAuthChallengeRequest(t *testing.T) {
	t.Parallel()

	t.Run("Given a malformed address", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail with a validation error on address", func(t *testing.T) {
			t.Parallel()

			request := &requests.GetAuthChallengeRequest{
				WalletAddressRequest: requests.WalletAddressRequest{Address: "0x1234"},
				ChainId:              80002,
			}

			var validationErr *apperrors.ValidationError
			err := request.ValidateAndSanitize()
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "address", validationErr.Field)
		})
	})

	t.Run("Given a blank chain id", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail with a validation error on chainId", func(t *testing.T) {
			t.Parallel()

			request := &requests.GetAuthChallengeRequest{
				WalletAddressRequest: requests.WalletAddressRequest{Address: validAddress},
			}

			var validationErr *apperrors.ValidationError
			err := request.ValidateAndSanitize()
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "chainId", validationErr.Field)
		})
	})

	t.Run("Given a fully populated request", func(t *testing.T) {
		t.Parallel()

		t.Run("Should pass validation", func(t *testing.T) {
			t.Parallel()

			request := &requests.GetAuthChallengeRequest{
				WalletAddressRequest: requests.WalletAddressRequest{Address: validAddress},
				ChainId:              80002,
			}

			assert.NoError(t, request.ValidateAndSanitize())
		})
	})
}
