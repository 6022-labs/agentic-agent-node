package models_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/stretchr/testify/assert"
)

func TestWhenReadingWhetherAPaymentReceiveRulePriceIsFree(t *testing.T) {
	t.Parallel()

	t.Run("Given a zero price", func(t *testing.T) {
		t.Parallel()

		t.Run("Should be free", func(t *testing.T) {
			t.Parallel()

			assert.True(t, (&models.PaymentReceiveRulePrice{Price: "0"}).IsFree())
		})
	})

	t.Run("Given a positive price", func(t *testing.T) {
		t.Parallel()

		t.Run("Should not be free", func(t *testing.T) {
			t.Parallel()

			assert.False(t, (&models.PaymentReceiveRulePrice{Price: "10000"}).IsFree())
		})
	})

	t.Run("Given a price that is not a number", func(t *testing.T) {
		t.Parallel()

		t.Run("Should not be free", func(t *testing.T) {
			t.Parallel()

			assert.False(t, (&models.PaymentReceiveRulePrice{Price: "free"}).IsFree())
		})
	})
}
