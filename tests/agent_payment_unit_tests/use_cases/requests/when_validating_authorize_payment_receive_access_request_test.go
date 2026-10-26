package requests_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingAuthorizePaymentReceiveAccessRequest(t *testing.T) {
	t.Parallel()

	t.Run("Given a whitespace-only address", func(t *testing.T) {
		t.Parallel()

		// It trims to empty, which means "no JWT was proven" rather than a bad address.
		t.Run("Should be accepted as no address at all", func(t *testing.T) {
			t.Parallel()

			request := &requests.AuthorizePaymentReceiveAccessRequest{Address: "   "}

			assert.NoError(t, request.ValidateAndSanitize())
			assert.Empty(t, request.Address)
		})
	})

	t.Run("Given an address that is not hex", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on the address field", func(t *testing.T) {
			t.Parallel()

			request := &requests.AuthorizePaymentReceiveAccessRequest{Address: "0xnope"}

			err := request.ValidateAndSanitize()

			var validationError *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationError)
			assert.Equal(t, "address", validationError.Field)
		})
	})

	t.Run("Given neither credential", func(t *testing.T) {
		t.Parallel()

		// A caller with nothing to show is valid input; the evaluator answers with a 402.
		t.Run("Should be accepted", func(t *testing.T) {
			t.Parallel()

			request := &requests.AuthorizePaymentReceiveAccessRequest{}

			assert.NoError(t, request.ValidateAndSanitize())
		})
	})

	t.Run("Given a padded address and signature", func(t *testing.T) {
		t.Parallel()

		t.Run("Should trim both", func(t *testing.T) {
			t.Parallel()

			address := common_tests.RandomEthAddress().Hex()
			request := &requests.AuthorizePaymentReceiveAccessRequest{
				Address:          "  " + address + "  ",
				PaymentSignature: "  a-header  ",
			}

			assert.NoError(t, request.ValidateAndSanitize())
			assert.Equal(t, address, request.Address)
			assert.Equal(t, "a-header", request.PaymentSignature)
		})
	})
}
