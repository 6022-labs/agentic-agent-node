package requests_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingSavePaymentReceiveRuleRequest(t *testing.T) {
	t.Parallel()

	generateSavePaymentReceiveRuleRequest := func() *requests.SavePaymentReceiveRuleRequest {
		return &requests.SavePaymentReceiveRuleRequest{
			SavePaymentRuleRequest: requests.SavePaymentRuleRequest{Name: "  Public price  "},
			Prices: []requests.SavePaymentReceiveRulePriceRequest{
				{Price: "0", AddressPatterns: []string{strings.ToLower(common_tests.RandomEthAddress().Hex())}},
				{Price: "10000", AddressPatterns: []string{".*"}},
			},
			Config: json.RawMessage(`{"network":"eip155:80002"}`),
		}
	}

	// The base request owns name and method; this proves it runs first.
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

	t.Run("Given the rule has no price", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError on the prices field", func(t *testing.T) {
			t.Parallel()

			req := generateSavePaymentReceiveRuleRequest()
			req.Prices = nil

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, req.ValidateAndSanitize(), &validationErr)
			assert.Equal(t, "prices", validationErr.Field)
		})
	})

	t.Run("Given a price row is invalid", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError on that row's field", func(t *testing.T) {
			t.Parallel()

			req := generateSavePaymentReceiveRuleRequest()
			req.Prices[1].Price = "-1"

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, req.ValidateAndSanitize(), &validationErr)
			assert.Equal(t, "prices[1].price", validationErr.Field)
		})
	})

	t.Run("Given two rows with the same price", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError on the prices field", func(t *testing.T) {
			t.Parallel()

			req := generateSavePaymentReceiveRuleRequest()
			req.Prices[1].Price = "00"

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, req.ValidateAndSanitize(), &validationErr)
			assert.Equal(t, "prices", validationErr.Field)
		})
	})

}
