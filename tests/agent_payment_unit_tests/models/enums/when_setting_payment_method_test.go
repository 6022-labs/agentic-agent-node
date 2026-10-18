package enums_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/stretchr/testify/assert"
)

func TestWhenSettingPaymentMethod(t *testing.T) {
	t.Parallel()

	t.Run("Given a known method in canonical casing", func(t *testing.T) {
		t.Parallel()
		var method enums.PaymentMethod

		t.Run("Should parse and assign it", func(t *testing.T) {
			t.Parallel()

			err := method.Set("x402")
			assert.NoError(t, err)
			assert.Equal(t, enums.PaymentMethodX402, method)
		})
	})

	t.Run("Given a known method in mixed casing with surrounding spaces", func(t *testing.T) {
		t.Parallel()
		var method enums.PaymentMethod

		t.Run("Should normalize and assign it", func(t *testing.T) {
			t.Parallel()

			err := method.Set(" X402 ")
			assert.NoError(t, err)
			assert.Equal(t, enums.PaymentMethodX402, method)
		})
	})

	t.Run("Given an unsupported method", func(t *testing.T) {
		t.Parallel()
		var method enums.PaymentMethod

		t.Run("Should reject with an error mentioning the value", func(t *testing.T) {
			t.Parallel()

			err := method.Set("lightning")
			assert.ErrorContains(t, err, "invalid PaymentMethod: lightning")
		})
	})

	t.Run("Given the full list of methods", func(t *testing.T) {
		t.Parallel()
		methods := enums.GetAllPaymentMethods()

		t.Run("Should contain every known method and exclude unknowns", func(t *testing.T) {
			t.Parallel()

			assert.True(t, methods.Contains(enums.PaymentMethodX402))
			assert.False(t, methods.Contains(enums.PaymentMethod("lightning")))
		})
	})
}
