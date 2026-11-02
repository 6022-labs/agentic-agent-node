package models_test

import (
	"math/big"
	"testing"

	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
)

func TestWhenCoveringAPaymentWithASpendRule(t *testing.T) {
	t.Parallel()

	t.Run("Given the rule has no per-payment cap", func(t *testing.T) {
		t.Parallel()

		t.Run("Should cover the payment", func(t *testing.T) {
			t.Parallel()

			rule := model_builders.NewPaymentSpendRuleBuilder().WithoutMaxAmountPerPayment().Build()

			covered, err := rule.CoversPayment(big.NewInt(1_000_000_000))

			assert.NoError(t, err)
			assert.True(t, covered)
		})
	})

	t.Run("Given the cap is not a usable amount", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return an error", func(t *testing.T) {
			t.Parallel()

			rule := model_builders.NewPaymentSpendRuleBuilder().WithMaxAmountPerPayment("not-a-number").Build()

			_, err := rule.CoversPayment(big.NewInt(1))

			assert.EqualError(t, err, "max amount per payment is not a usable amount")
		})
	})

	t.Run("Given the cap is negative", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return an error", func(t *testing.T) {
			t.Parallel()

			rule := model_builders.NewPaymentSpendRuleBuilder().WithMaxAmountPerPayment("-1").Build()

			_, err := rule.CoversPayment(big.NewInt(1))

			assert.EqualError(t, err, "max amount per payment is not a usable amount")
		})
	})

	t.Run("Given the amount equals the cap", func(t *testing.T) {
		t.Parallel()

		t.Run("Should cover the payment", func(t *testing.T) {
			t.Parallel()

			rule := model_builders.NewPaymentSpendRuleBuilder().WithMaxAmountPerPayment("1000").Build()

			covered, err := rule.CoversPayment(big.NewInt(1000))

			assert.NoError(t, err)
			assert.True(t, covered)
		})
	})

	t.Run("Given the amount is above the cap", func(t *testing.T) {
		t.Parallel()

		t.Run("Should not cover the payment", func(t *testing.T) {
			t.Parallel()

			rule := model_builders.NewPaymentSpendRuleBuilder().WithMaxAmountPerPayment("1000").Build()

			covered, err := rule.CoversPayment(big.NewInt(1001))

			assert.NoError(t, err)
			assert.False(t, covered)
		})
	})
}
