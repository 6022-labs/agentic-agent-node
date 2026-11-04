package requests_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingUpdatePaymentSpendPolicyRequest(t *testing.T) {
	t.Parallel()

	t.Run("Given a payee pattern that is not a valid regex", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on the payToPatterns field", func(t *testing.T) {
			t.Parallel()

			request := &requests.UpdatePaymentSpendPolicyRequest{PayToPatterns: []string{"0x[("}}

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, request.ValidateAndSanitize(), &validationErr)
			assert.Equal(t, "payToPatterns", validationErr.Field)
		})
	})

	t.Run("Given no payee at all", func(t *testing.T) {
		t.Parallel()

		// An empty list is how an owner stops every payment without deleting the rules.
		t.Run("Should accept it", func(t *testing.T) {
			t.Parallel()

			request := &requests.UpdatePaymentSpendPolicyRequest{PayToPatterns: []string{}}

			assert.NoError(t, request.ValidateAndSanitize())
		})
	})
}
