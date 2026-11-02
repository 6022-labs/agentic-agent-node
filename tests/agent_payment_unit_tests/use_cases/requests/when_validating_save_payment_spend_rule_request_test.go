package requests_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingSavePaymentSpendRuleRequest(t *testing.T) {
	t.Parallel()

	generateSavePaymentSpendRuleRequest := func(maxAmountPerPayment string) *requests.SavePaymentSpendRuleRequest {
		return &requests.SavePaymentSpendRuleRequest{
			Name:                "  Test token  ",
			MaxAmountPerPayment: &maxAmountPerPayment,
			Windows: []requests.SavePaymentSpendRuleWindowRequest{
				{WindowSeconds: 86400, MaxAmount: "50000"},
			},
			Config: json.RawMessage(`{"network":"eip155:8453"}`),
		}
	}

	t.Run("Given the name is empty", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on the name field", func(t *testing.T) {
			t.Parallel()

			request := generateSavePaymentSpendRuleRequest("10000")
			request.Name = "   "

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, request.ValidateAndSanitize(), &validationErr)
			assert.Equal(t, "name", validationErr.Field)
		})
	})

	t.Run("Given the name is longer than 64 characters", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on the name field", func(t *testing.T) {
			t.Parallel()

			request := generateSavePaymentSpendRuleRequest("10000")
			request.Name = strings.Repeat("a", 65)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, request.ValidateAndSanitize(), &validationErr)
			assert.Equal(t, "name", validationErr.Field)
		})
	})

	t.Run("Given the per-payment cap is blank", func(t *testing.T) {
		t.Parallel()

		t.Run("Should treat it as no cap", func(t *testing.T) {
			t.Parallel()

			request := generateSavePaymentSpendRuleRequest("  ")

			assert.NoError(t, request.ValidateAndSanitize())
			assert.Nil(t, request.MaxAmountPerPayment)
		})
	})

	t.Run("Given the per-payment cap is negative", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on the per-payment field", func(t *testing.T) {
			t.Parallel()

			request := generateSavePaymentSpendRuleRequest("-1")

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, request.ValidateAndSanitize(), &validationErr)
			assert.Equal(t, "maxAmountPerPayment", validationErr.Field)
		})
	})

	t.Run("Given the per-payment cap is not an integer", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on the per-payment field", func(t *testing.T) {
			t.Parallel()

			request := generateSavePaymentSpendRuleRequest("1.5")

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, request.ValidateAndSanitize(), &validationErr)
			assert.Equal(t, "maxAmountPerPayment", validationErr.Field)
		})
	})

	t.Run("Given the per-payment cap is above the largest uint256", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on the per-payment field", func(t *testing.T) {
			t.Parallel()

			request := generateSavePaymentSpendRuleRequest(
				"115792089237316195423570985008687907853269984665640564039457584007913129639936",
			)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, request.ValidateAndSanitize(), &validationErr)
			assert.Equal(t, "maxAmountPerPayment", validationErr.Field)
		})
	})

	t.Run("Given neither a per-payment cap nor a window", func(t *testing.T) {
		t.Parallel()

		// Nothing is unlimited: a rule with no bound at all would let the agent pay anything.
		t.Run("Should return a validation error on the windows field", func(t *testing.T) {
			t.Parallel()

			request := generateSavePaymentSpendRuleRequest("")
			request.Windows = nil

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, request.ValidateAndSanitize(), &validationErr)
			assert.Equal(t, "windows", validationErr.Field)
		})
	})

	t.Run("Given a window is invalid", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return the window's validation error", func(t *testing.T) {
			t.Parallel()

			request := generateSavePaymentSpendRuleRequest("10000")
			request.Windows = append(request.Windows, requests.SavePaymentSpendRuleWindowRequest{MaxAmount: "1"})

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, request.ValidateAndSanitize(), &validationErr)
			assert.Equal(t, "windows[1].windowSeconds", validationErr.Field)
		})
	})

	t.Run("Given two windows share a period", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on the windows field", func(t *testing.T) {
			t.Parallel()

			request := generateSavePaymentSpendRuleRequest("10000")
			request.Windows = append(
				request.Windows,
				requests.SavePaymentSpendRuleWindowRequest{WindowSeconds: 86400, MaxAmount: "60000"},
			)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, request.ValidateAndSanitize(), &validationErr)
			assert.Equal(t, "windows", validationErr.Field)
		})
	})

	t.Run("Given a per-payment cap above a window budget", func(t *testing.T) {
		t.Parallel()

		// No single payment could ever fit, so the cap would be dead config.
		t.Run("Should return a validation error on the per-payment field", func(t *testing.T) {
			t.Parallel()

			request := generateSavePaymentSpendRuleRequest("50001")

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, request.ValidateAndSanitize(), &validationErr)
			assert.Equal(t, "maxAmountPerPayment", validationErr.Field)
		})
	})

	t.Run("Given an unsupported method", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on the method field", func(t *testing.T) {
			t.Parallel()

			request := generateSavePaymentSpendRuleRequest("10000")
			request.Method = "lightning"

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, request.ValidateAndSanitize(), &validationErr)
			assert.Equal(t, "method", validationErr.Field)
		})
	})

	t.Run("Given an explicit supported method", func(t *testing.T) {
		t.Parallel()

		t.Run("Should normalize it", func(t *testing.T) {
			t.Parallel()

			request := generateSavePaymentSpendRuleRequest("10000")
			request.Method = " X402 "

			assert.NoError(t, request.ValidateAndSanitize())
			assert.Equal(t, enums.PaymentMethodX402.String(), request.Method)
		})
	})

	t.Run("Given a per-payment cap equal to a window budget", func(t *testing.T) {
		t.Parallel()

		t.Run("Should accept it", func(t *testing.T) {
			t.Parallel()

			request := generateSavePaymentSpendRuleRequest("50000")

			assert.NoError(t, request.ValidateAndSanitize())
		})
	})

	t.Run("Given a per-payment cap and no window", func(t *testing.T) {
		t.Parallel()

		t.Run("Should accept it", func(t *testing.T) {
			t.Parallel()

			request := generateSavePaymentSpendRuleRequest("10000")
			request.Windows = nil

			assert.NoError(t, request.ValidateAndSanitize())
		})
	})

	t.Run("Given a zero per-payment cap", func(t *testing.T) {
		t.Parallel()

		// "0" is the deny value, so it must survive validation rather than be read as blank.
		t.Run("Should keep it as a cap", func(t *testing.T) {
			t.Parallel()

			request := generateSavePaymentSpendRuleRequest("0")

			assert.NoError(t, request.ValidateAndSanitize())
			if assert.NotNil(t, request.MaxAmountPerPayment) {
				assert.Equal(t, "0", *request.MaxAmountPerPayment)
			}
		})
	})

	t.Run("Given the request is valid", func(t *testing.T) {
		t.Parallel()

		t.Run("Should trim the name", func(t *testing.T) {
			t.Parallel()

			request := generateSavePaymentSpendRuleRequest("10000")

			assert.NoError(t, request.ValidateAndSanitize())
			assert.Equal(t, "Test token", request.Name)
		})

		t.Run("Should default the method to x402", func(t *testing.T) {
			t.Parallel()

			request := generateSavePaymentSpendRuleRequest("10000")

			assert.NoError(t, request.ValidateAndSanitize())
			assert.Equal(t, enums.PaymentMethodX402.String(), request.Method)
		})
	})
}
