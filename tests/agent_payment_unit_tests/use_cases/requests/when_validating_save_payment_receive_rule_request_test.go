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

func TestWhenValidatingSavePaymentReceiveRuleRequest(t *testing.T) {
	t.Parallel()

	generateSavePaymentReceiveRuleRequest := func() *requests.SavePaymentReceiveRuleRequest {
		return &requests.SavePaymentReceiveRuleRequest{
			Name:   "  Public price  ",
			Price:  "10000",
			Config: json.RawMessage(`{"network":"eip155:80002"}`),
		}
	}

	t.Run("Given a blank rule name", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError on the name field", func(t *testing.T) {
			t.Parallel()

			req := generateSavePaymentReceiveRuleRequest()
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

			req := generateSavePaymentReceiveRuleRequest()
			req.Name = strings.Repeat("a", 65)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, req.ValidateAndSanitize(), &validationErr)
			assert.Equal(t, "name", validationErr.Field)
		})
	})

	t.Run("Given the rule has a non-numeric price", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError on the price field", func(t *testing.T) {
			t.Parallel()

			req := generateSavePaymentReceiveRuleRequest()
			req.Price = "abc"

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, req.ValidateAndSanitize(), &validationErr)
			assert.Equal(t, "price", validationErr.Field)
		})
	})

	t.Run("Given the rule has a negative price", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError on the price field", func(t *testing.T) {
			t.Parallel()

			req := generateSavePaymentReceiveRuleRequest()
			req.Price = "-1"

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, req.ValidateAndSanitize(), &validationErr)
			assert.Equal(t, "price", validationErr.Field)
		})
	})

	t.Run("Given an unsupported method", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError on the method field", func(t *testing.T) {
			t.Parallel()

			req := generateSavePaymentReceiveRuleRequest()
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

			req := generateSavePaymentReceiveRuleRequest()

			assert.NoError(t, req.ValidateAndSanitize())
			assert.Equal(t, "Public price", req.Name)
		})
	})

	t.Run("Given no method", func(t *testing.T) {
		t.Parallel()

		t.Run("Should default the method to x402", func(t *testing.T) {
			t.Parallel()

			req := generateSavePaymentReceiveRuleRequest()

			assert.NoError(t, req.ValidateAndSanitize())
			assert.Equal(t, enums.PaymentMethodX402.String(), req.Method)
		})
	})

	t.Run("Given an explicit supported method", func(t *testing.T) {
		t.Parallel()

		t.Run("Should normalize it", func(t *testing.T) {
			t.Parallel()

			req := generateSavePaymentReceiveRuleRequest()
			req.Method = " X402 "

			assert.NoError(t, req.ValidateAndSanitize())
			assert.Equal(t, enums.PaymentMethodX402.String(), req.Method)
		})
	})

	t.Run("Given the rule is free", func(t *testing.T) {
		t.Parallel()

		t.Run("Should accept a zero price", func(t *testing.T) {
			t.Parallel()

			req := generateSavePaymentReceiveRuleRequest()
			req.Price = "0"

			assert.NoError(t, req.ValidateAndSanitize())
		})
	})
}
