package requests_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
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
