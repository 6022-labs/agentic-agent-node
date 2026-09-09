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

	generateSavePaymentSpendRuleRequest := func() *requests.SavePaymentSpendRuleRequest {
		return &requests.SavePaymentSpendRuleRequest{
			Name:                "  Anything, unbounded  ",
			MaxAmountPerPayment: "10000",
			Config:              json.RawMessage(`{"networkPatterns":[".*"]}`),
		}
	}

	t.Run("Given the request is valid", func(t *testing.T) {
		t.Parallel()

		t.Run("Should trim the name and default the method to x402", func(t *testing.T) {
			t.Parallel()

			request := generateSavePaymentSpendRuleRequest()
			err := request.ValidateAndSanitize()

			assert.NoError(t, err)
			assert.Equal(t, "Anything, unbounded", request.Name)
			assert.Equal(t, enums.PaymentMethodX402.String(), request.Method)
		})
	})

	t.Run("Given an explicit supported method", func(t *testing.T) {
		t.Parallel()

		t.Run("Should normalize it", func(t *testing.T) {
			t.Parallel()

			request := generateSavePaymentSpendRuleRequest()
			request.Method = " X402 "

			assert.NoError(t, request.ValidateAndSanitize())
			assert.Equal(t, enums.PaymentMethodX402.String(), request.Method)
		})
	})

	t.Run("Given a period with a budget that covers the per-payment cap", func(t *testing.T) {
		t.Parallel()

		t.Run("Should accept both bounds", func(t *testing.T) {
			t.Parallel()

			request := generateSavePaymentSpendRuleRequest()
			request.WindowSeconds = 86400
			request.MaxAmountPerWindow = "  50000  "

			assert.NoError(t, request.ValidateAndSanitize())
			assert.Equal(t, "50000", request.MaxAmountPerWindow)
			assert.Equal(t, 86400, request.WindowSeconds)
		})
	})

	t.Run("Given no period at all", func(t *testing.T) {
		t.Parallel()

		// The column is NOT NULL, so an absent budget still has to reach the row as "0".
		t.Run("Should default the budget to the deny value", func(t *testing.T) {
			t.Parallel()

			request := generateSavePaymentSpendRuleRequest()

			assert.NoError(t, request.ValidateAndSanitize())
			assert.Equal(t, "0", request.MaxAmountPerWindow)
		})
	})

	t.Run("Given a budget with no period to spend it over", func(t *testing.T) {
		t.Parallel()

		// Accepting it would silently drop the bound the owner asked for.
		t.Run("Should return a validation error on the budget field", func(t *testing.T) {
			t.Parallel()

			request := generateSavePaymentSpendRuleRequest()
			request.MaxAmountPerWindow = "50000"

			err := request.ValidateAndSanitize()

			var validationError *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationError)
			assert.Equal(t, "maxAmountPerWindow", validationError.Field)
		})
	})

	t.Run("Given a period with no budget", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on the budget field", func(t *testing.T) {
			t.Parallel()

			request := generateSavePaymentSpendRuleRequest()
			request.WindowSeconds = 86400

			err := request.ValidateAndSanitize()

			var validationError *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationError)
			assert.Equal(t, "maxAmountPerWindow", validationError.Field)
		})
	})

	t.Run("Given a negative period", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on the period field", func(t *testing.T) {
			t.Parallel()

			request := generateSavePaymentSpendRuleRequest()
			request.WindowSeconds = -1

			err := request.ValidateAndSanitize()

			var validationError *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationError)
			assert.Equal(t, "windowSeconds", validationError.Field)
		})
	})

	t.Run("Given a per-payment cap above the period budget", func(t *testing.T) {
		t.Parallel()

		// No payment could ever fit the period, so the rule would deny everything it matches.
		t.Run("Should return a validation error on the per-payment field", func(t *testing.T) {
			t.Parallel()

			request := generateSavePaymentSpendRuleRequest()
			request.MaxAmountPerPayment = "10000"
			request.WindowSeconds = 86400
			request.MaxAmountPerWindow = "9999"

			err := request.ValidateAndSanitize()

			var validationError *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationError)
			assert.Equal(t, "maxAmountPerPayment", validationError.Field)
		})
	})

	t.Run("Given a period budget that is not a usable amount", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on the budget field", func(t *testing.T) {
			t.Parallel()

			request := generateSavePaymentSpendRuleRequest()
			request.WindowSeconds = 86400
			request.MaxAmountPerWindow = "-1"

			err := request.ValidateAndSanitize()

			var validationError *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationError)
			assert.Equal(t, "maxAmountPerWindow", validationError.Field)
		})
	})

	t.Run("Given an unsupported method", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on the method field", func(t *testing.T) {
			t.Parallel()

			request := generateSavePaymentSpendRuleRequest()
			request.Method = "lightning"

			err := request.ValidateAndSanitize()

			assert.Error(t, err)
		})
	})

	t.Run("Given the name is empty", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error", func(t *testing.T) {
			t.Parallel()

			request := generateSavePaymentSpendRuleRequest()
			request.Name = "   "

			assert.Error(t, request.ValidateAndSanitize())
		})
	})

	t.Run("Given the name is longer than 64 characters", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error", func(t *testing.T) {
			t.Parallel()

			request := generateSavePaymentSpendRuleRequest()
			request.Name = strings.Repeat("a", 65)

			assert.Error(t, request.ValidateAndSanitize())
		})
	})

	t.Run("Given the max amount is zero", func(t *testing.T) {
		t.Parallel()

		// "0" is the deny idiom, not an invalid cap.
		t.Run("Should accept it", func(t *testing.T) {
			t.Parallel()

			request := generateSavePaymentSpendRuleRequest()
			request.MaxAmountPerPayment = "0"

			assert.NoError(t, request.ValidateAndSanitize())
		})
	})

	t.Run("Given the max amount is negative", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on the max amount field", func(t *testing.T) {
			t.Parallel()

			request := generateSavePaymentSpendRuleRequest()
			request.MaxAmountPerPayment = "-1"

			assert.Error(t, request.ValidateAndSanitize())
		})
	})

	t.Run("Given the max amount is not an integer", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error", func(t *testing.T) {
			t.Parallel()

			request := generateSavePaymentSpendRuleRequest()
			request.MaxAmountPerPayment = "1.5"

			assert.Error(t, request.ValidateAndSanitize())
		})
	})

	t.Run("Given the max amount is blank", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error", func(t *testing.T) {
			t.Parallel()

			request := generateSavePaymentSpendRuleRequest()
			request.MaxAmountPerPayment = "  "

			assert.Error(t, request.ValidateAndSanitize())
		})
	})
}
