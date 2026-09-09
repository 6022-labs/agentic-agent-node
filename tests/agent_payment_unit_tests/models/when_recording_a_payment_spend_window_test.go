package models_test

import (
	"math/big"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
)

const oneDay = 24 * 60 * 60

func TestWhenRecordingAPaymentSpendWindow(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	t.Run("Given the rule has no window", func(t *testing.T) {
		t.Parallel()

		t.Run("Should record nothing and allow the payment", func(t *testing.T) {
			t.Parallel()

			rule := model_builders.NewPaymentSpendRuleBuilder().Build()

			recorded, err := rule.RecordWindowSpend(big.NewInt(1000), now)

			assert.NoError(t, err)
			assert.True(t, recorded)
			assert.Equal(t, "0", rule.WindowSpent)
		})
	})

	t.Run("Given a daily budget with room left", func(t *testing.T) {
		t.Parallel()

		t.Run("Should accumulate across payments", func(t *testing.T) {
			t.Parallel()

			rule := model_builders.NewPaymentSpendRuleBuilder().Build()
			rule.WindowSeconds = oneDay
			rule.MaxAmountPerWindow = "1000"

			for range 3 {
				recorded, err := rule.RecordWindowSpend(big.NewInt(300), now)
				assert.NoError(t, err)
				assert.True(t, recorded)
			}

			assert.Equal(t, "900", rule.WindowSpent)
		})
	})

	t.Run("Given the payment would exceed the period", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse without recording it", func(t *testing.T) {
			t.Parallel()

			rule := model_builders.NewPaymentSpendRuleBuilder().Build()
			rule.WindowSeconds = oneDay
			rule.MaxAmountPerWindow = "1000"

			recorded, err := rule.RecordWindowSpend(big.NewInt(900), now)
			assert.NoError(t, err)
			assert.True(t, recorded)

			recorded, err = rule.RecordWindowSpend(big.NewInt(200), now)

			assert.NoError(t, err)
			assert.False(t, recorded)
			assert.Equal(t, "900", rule.WindowSpent)
		})
	})

	t.Run("Given the period has elapsed", func(t *testing.T) {
		t.Parallel()

		// The budget returns on its own; nothing needs resetting by hand.
		t.Run("Should start over on the next payment", func(t *testing.T) {
			t.Parallel()

			rule := model_builders.NewPaymentSpendRuleBuilder().Build()
			rule.WindowSeconds = oneDay
			rule.MaxAmountPerWindow = "1000"

			recorded, err := rule.RecordWindowSpend(big.NewInt(1000), now)
			assert.NoError(t, err)
			assert.True(t, recorded)

			refused, err := rule.RecordWindowSpend(big.NewInt(1), now)
			assert.NoError(t, err)
			assert.False(t, refused)

			recorded, err = rule.RecordWindowSpend(big.NewInt(1000), now.Add(25*time.Hour))

			assert.NoError(t, err)
			assert.True(t, recorded)
			assert.Equal(t, "1000", rule.WindowSpent)
		})
	})

	t.Run("Given a stored amount is not usable", func(t *testing.T) {
		t.Parallel()

		for _, unusable := range []string{"not-a-number", "-1"} {
			t.Run("Should fail loudly for a budget of "+unusable, func(t *testing.T) {
				t.Parallel()

				rule := model_builders.NewPaymentSpendRuleBuilder().Build()
				rule.WindowSeconds = oneDay
				rule.MaxAmountPerWindow = unusable

				recorded, err := rule.RecordWindowSpend(big.NewInt(1), now)

				assert.Error(t, err)
				assert.False(t, recorded)
			})
		}
	})

	t.Run("Given the per-payment cap", func(t *testing.T) {
		t.Parallel()

		t.Run("Should allow an amount equal to it and refuse anything above", func(t *testing.T) {
			t.Parallel()

			rule := model_builders.NewPaymentSpendRuleBuilder().Build()
			rule.MaxAmountPerPayment = "1000"

			covered, err := rule.CoversPayment(big.NewInt(1000))
			assert.NoError(t, err)
			assert.True(t, covered)

			covered, err = rule.CoversPayment(big.NewInt(1001))
			assert.NoError(t, err)
			assert.False(t, covered)
		})
	})
}
