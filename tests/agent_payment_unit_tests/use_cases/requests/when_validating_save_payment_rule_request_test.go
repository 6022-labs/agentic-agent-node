package requests_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingSavePaymentRuleRequest(t *testing.T) {
	t.Parallel()

	generateSavePaymentRuleRequest := func() *requests.SavePaymentRuleRequest {
		return &requests.SavePaymentRuleRequest{
			Name:   "  Public price  ",
			Price:  "10000",
			Config: json.RawMessage(`{"network":"eip155:80002"}`),
		}
	}

	t.Run("Given the request is valid", func(t *testing.T) {
		t.Parallel()

		t.Run("Should trim the name and default the method to x402", func(t *testing.T) {
			t.Parallel()

			req := generateSavePaymentRuleRequest()
			err := req.ValidateAndSanitize()

			assert.NoError(t, err)
			assert.Equal(t, "Public price", req.Name)
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

	t.Run("Given an unsupported method", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error", func(t *testing.T) {
			t.Parallel()

			req := generateSavePaymentRuleRequest()
			req.Method = "lightning"

			assert.Error(t, req.ValidateAndSanitize())
		})
	})

	t.Run("Given the rule has no name", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error", func(t *testing.T) {
			t.Parallel()

			req := generateSavePaymentRuleRequest()
			req.Name = "   "

			assert.Error(t, req.ValidateAndSanitize())
		})
	})

	t.Run("Given the rule name exceeds 64 characters", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error", func(t *testing.T) {
			t.Parallel()

			req := generateSavePaymentRuleRequest()
			req.Name = strings.Repeat("a", 65)

			assert.Error(t, req.ValidateAndSanitize())
		})
	})

	t.Run("Given the rule is free", func(t *testing.T) {
		t.Parallel()

		t.Run("Should accept a zero price", func(t *testing.T) {
			t.Parallel()

			req := generateSavePaymentRuleRequest()
			req.Price = "0"

			assert.NoError(t, req.ValidateAndSanitize())
		})
	})

	t.Run("Given the rule has a negative price", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error", func(t *testing.T) {
			t.Parallel()

			req := generateSavePaymentRuleRequest()
			req.Price = "-1"

			assert.Error(t, req.ValidateAndSanitize())
		})
	})

	t.Run("Given the rule has a non-numeric price", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error", func(t *testing.T) {
			t.Parallel()

			req := generateSavePaymentRuleRequest()
			req.Price = "abc"

			assert.Error(t, req.ValidateAndSanitize())
		})
	})
}
