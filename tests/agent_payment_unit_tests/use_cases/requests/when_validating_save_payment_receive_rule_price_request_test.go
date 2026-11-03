package requests_test

import (
	"strings"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingSavePaymentReceiveRulePriceRequest(t *testing.T) {
	t.Parallel()

	generateSavePaymentReceiveRulePriceRequest := func() *requests.SavePaymentReceiveRulePriceRequest {
		return &requests.SavePaymentReceiveRulePriceRequest{
			Price:           " 10000 ",
			AddressPatterns: []string{".*"},
		}
	}

	t.Run("Given a non-numeric price", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError on the row's price field", func(t *testing.T) {
			t.Parallel()

			req := generateSavePaymentReceiveRulePriceRequest()
			req.Price = "abc"

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, req.ValidateAndSanitize(2), &validationErr)
			assert.Equal(t, "prices[2].price", validationErr.Field)
		})
	})

	t.Run("Given a negative price", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError on the row's price field", func(t *testing.T) {
			t.Parallel()

			req := generateSavePaymentReceiveRulePriceRequest()
			req.Price = "-1"

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, req.ValidateAndSanitize(0), &validationErr)
			assert.Equal(t, "prices[0].price", validationErr.Field)
		})
	})

	t.Run("Given a price above 2^256-1", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError on the row's price field", func(t *testing.T) {
			t.Parallel()

			req := generateSavePaymentReceiveRulePriceRequest()
			req.Price = "1" + strings.Repeat("0", 78)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, req.ValidateAndSanitize(0), &validationErr)
			assert.Equal(t, "prices[0].price", validationErr.Field)
		})
	})

	t.Run("Given no address pattern", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError on the row's addressPatterns field", func(t *testing.T) {
			t.Parallel()

			req := generateSavePaymentReceiveRulePriceRequest()
			req.AddressPatterns = nil

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, req.ValidateAndSanitize(1), &validationErr)
			assert.Equal(t, "prices[1].addressPatterns", validationErr.Field)
		})
	})

	t.Run("Given an address pattern that is not a valid regex", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError on the row's addressPatterns field", func(t *testing.T) {
			t.Parallel()

			req := generateSavePaymentReceiveRulePriceRequest()
			req.AddressPatterns = []string{"0x[("}

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, req.ValidateAndSanitize(0), &validationErr)
			assert.Equal(t, "prices[0].addressPatterns", validationErr.Field)
		})
	})

	t.Run("Given a price with leading zeros", func(t *testing.T) {
		t.Parallel()

		t.Run("Should store it without the leading zeros", func(t *testing.T) {
			t.Parallel()

			req := generateSavePaymentReceiveRulePriceRequest()
			req.Price = " 007 "

			assert.NoError(t, req.ValidateAndSanitize(0))
			assert.Equal(t, "7", req.Price)
		})
	})
}
