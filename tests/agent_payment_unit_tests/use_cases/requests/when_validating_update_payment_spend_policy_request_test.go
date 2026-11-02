package requests_test

import (
	"strings"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	common_tests "github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingUpdatePaymentSpendPolicyRequest(t *testing.T) {
	t.Parallel()

	t.Run("Given a blank payee pattern", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on the payToPatterns field", func(t *testing.T) {
			t.Parallel()

			request := &requests.UpdatePaymentSpendPolicyRequest{PayToPatterns: []string{".*", "  "}}

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, request.ValidateAndSanitize(), &validationErr)
			assert.Equal(t, "payToPatterns", validationErr.Field)
		})
	})

	t.Run("Given a payee pattern that cannot compile", func(t *testing.T) {
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

	t.Run("Given a payee pattern spelled as a checksummed address", func(t *testing.T) {
		t.Parallel()

		t.Run("Should be stored lowercased", func(t *testing.T) {
			t.Parallel()

			payTo := common_tests.RandomEthAddress().Hex()
			request := &requests.UpdatePaymentSpendPolicyRequest{PayToPatterns: []string{payTo}}

			assert.NoError(t, request.ValidateAndSanitize())
			assert.Equal(t, []string{strings.ToLower(payTo)}, request.PayToPatterns)
		})
	})

	t.Run("Given the same payee spelled checksummed and lowercased", func(t *testing.T) {
		t.Parallel()

		t.Run("Should keep it once", func(t *testing.T) {
			t.Parallel()

			payTo := common_tests.RandomEthAddress().Hex()
			request := &requests.UpdatePaymentSpendPolicyRequest{PayToPatterns: []string{payTo, strings.ToLower(payTo)}}

			assert.NoError(t, request.ValidateAndSanitize())
			assert.Equal(t, []string{strings.ToLower(payTo)}, request.PayToPatterns)
		})
	})

	t.Run("Given a payee pattern with surrounding spaces", func(t *testing.T) {
		t.Parallel()

		t.Run("Should trim it", func(t *testing.T) {
			t.Parallel()

			request := &requests.UpdatePaymentSpendPolicyRequest{PayToPatterns: []string{"  0xabc.*  "}}

			assert.NoError(t, request.ValidateAndSanitize())
			assert.Equal(t, []string{"0xabc.*"}, request.PayToPatterns)
		})
	})
}
