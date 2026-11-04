package requests_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingSavePaymentSpendRuleWindowRequest(t *testing.T) {
	t.Parallel()

	t.Run("Given the period is zero", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on the period field", func(t *testing.T) {
			t.Parallel()

			request := &requests.SavePaymentSpendRuleWindowRequest{MaxAmount: "50000"}

			_, err := request.ValidateAndSanitize(2)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "windows[2].windowSeconds", validationErr.Field)
		})
	})

	t.Run("Given a negative period", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on the period field", func(t *testing.T) {
			t.Parallel()

			request := &requests.SavePaymentSpendRuleWindowRequest{WindowSeconds: -1, MaxAmount: "50000"}

			_, err := request.ValidateAndSanitize(0)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "windows[0].windowSeconds", validationErr.Field)
		})
	})

	// A longer period would overflow time.Duration and never accumulate spend.
	t.Run("Given the period is longer than 10 years", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on the period field", func(t *testing.T) {
			t.Parallel()

			request := &requests.SavePaymentSpendRuleWindowRequest{
				WindowSeconds: requests.MaxPaymentSpendWindowSeconds + 1,
				MaxAmount:     "50000",
			}

			_, err := request.ValidateAndSanitize(0)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "windows[0].windowSeconds", validationErr.Field)
		})
	})

	t.Run("Given the budget is blank", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on the budget field", func(t *testing.T) {
			t.Parallel()

			request := &requests.SavePaymentSpendRuleWindowRequest{WindowSeconds: 86400, MaxAmount: " "}

			_, err := request.ValidateAndSanitize(0)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "windows[0].maxAmount", validationErr.Field)
		})
	})

	t.Run("Given the budget is negative", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on the budget field", func(t *testing.T) {
			t.Parallel()

			request := &requests.SavePaymentSpendRuleWindowRequest{WindowSeconds: 86400, MaxAmount: "-1"}

			_, err := request.ValidateAndSanitize(0)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "windows[0].maxAmount", validationErr.Field)
		})
	})

	t.Run("Given the budget is not an integer", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on the budget field", func(t *testing.T) {
			t.Parallel()

			request := &requests.SavePaymentSpendRuleWindowRequest{WindowSeconds: 86400, MaxAmount: "1.5"}

			_, err := request.ValidateAndSanitize(0)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "windows[0].maxAmount", validationErr.Field)
		})
	})

	t.Run("Given the budget is above the largest uint256", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error on the budget field", func(t *testing.T) {
			t.Parallel()

			request := &requests.SavePaymentSpendRuleWindowRequest{
				WindowSeconds: 86400,
				MaxAmount:     "115792089237316195423570985008687907853269984665640564039457584007913129639936",
			}

			_, err := request.ValidateAndSanitize(0)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "windows[0].maxAmount", validationErr.Field)
		})
	})

	t.Run("Given a budget with leading zeros", func(t *testing.T) {
		t.Parallel()

		t.Run("Should store it without the leading zeros", func(t *testing.T) {
			t.Parallel()

			request := &requests.SavePaymentSpendRuleWindowRequest{WindowSeconds: 86400, MaxAmount: "0050000"}

			_, err := request.ValidateAndSanitize(0)

			assert.NoError(t, err)
			assert.Equal(t, "50000", request.MaxAmount)
		})
	})

	t.Run("Given the window is valid", func(t *testing.T) {
		t.Parallel()

		t.Run("Should trim the budget", func(t *testing.T) {
			t.Parallel()

			request := &requests.SavePaymentSpendRuleWindowRequest{WindowSeconds: 86400, MaxAmount: " 50000 "}

			_, err := request.ValidateAndSanitize(0)

			assert.NoError(t, err)
			assert.Equal(t, "50000", request.MaxAmount)
		})

		t.Run("Should return the parsed budget", func(t *testing.T) {
			t.Parallel()

			request := &requests.SavePaymentSpendRuleWindowRequest{WindowSeconds: 86400, MaxAmount: "50000"}

			maxAmount, err := request.ValidateAndSanitize(0)

			assert.NoError(t, err)
			assert.Equal(t, int64(50000), maxAmount.Int64())
		})
	})
}
