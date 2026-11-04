package requests_test

import (
	"strings"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingSavePaymentRuleRequest(t *testing.T) {
	t.Parallel()

	generateSavePaymentRuleRequest := func() *requests.SavePaymentRuleRequest {
		return &requests.SavePaymentRuleRequest{Name: "  Public price  "}
	}

	t.Run("Given a blank rule name", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError on the name field", func(t *testing.T) {
			t.Parallel()

			req := generateSavePaymentRuleRequest()
			req.Name = "   "

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, req.ValidateAndSanitize(), &validationErr)
			assert.Equal(t, "name", validationErr.Field)
		})
	})

	t.Run("Given a rule name longer than 64 characters", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError on the name field", func(t *testing.T) {
			t.Parallel()

			req := generateSavePaymentRuleRequest()
			req.Name = strings.Repeat("a", 65)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, req.ValidateAndSanitize(), &validationErr)
			assert.Equal(t, "name", validationErr.Field)
		})
	})

	t.Run("Given an unsupported method", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError on the method field", func(t *testing.T) {
			t.Parallel()

			req := generateSavePaymentRuleRequest()
			req.Method = "lightning"

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, req.ValidateAndSanitize(), &validationErr)
			assert.Equal(t, "method", validationErr.Field)
		})
	})

	t.Run("Given a padded rule name", func(t *testing.T) {
		t.Parallel()

		t.Run("Should trim the name", func(t *testing.T) {
			t.Parallel()

			req := generateSavePaymentRuleRequest()

			assert.NoError(t, req.ValidateAndSanitize())
			assert.Equal(t, "Public price", req.Name)
		})
	})

	t.Run("Given no method", func(t *testing.T) {
		t.Parallel()

		t.Run("Should default the method to x402", func(t *testing.T) {
			t.Parallel()

			req := generateSavePaymentRuleRequest()

			assert.NoError(t, req.ValidateAndSanitize())
			assert.Equal(t, enums.PaymentMethodX402.String(), req.Method)
		})
	})

	t.Run("Given an explicit supported method", func(t *testing.T) {
		t.Parallel()

		t.Run("Should normalize it", func(t *testing.T) {
			t.Parallel()

			req := generateSavePaymentRuleRequest()
			req.Method = " X402 "

			assert.NoError(t, req.ValidateAndSanitize())
			assert.Equal(t, enums.PaymentMethodX402.String(), req.Method)
		})
	})
}
