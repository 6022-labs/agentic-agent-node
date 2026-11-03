package models_test

import (
	"math/big"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
)

func TestWhenRecordingAPaymentSpendWindow(t *testing.T) {
	t.Parallel()

	// Sub-second on purpose: a period starts on a whole second so it survives every driver's datetime precision.
	now := time.Date(2026, 10, 28, 12, 0, 0, 750_000_000, time.UTC)

	t.Run("Given the budget is not a usable amount", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse the budget as unusable", func(t *testing.T) {
			t.Parallel()

			window := model_builders.NewPaymentSpendRuleWindowBuilder().WithMaxAmount("not-a-number").Build()

			_, err := window.RecordSpend(big.NewInt(1), now)

			assert.EqualError(t, err, "window budget is not a usable amount")
		})
	})

	t.Run("Given the budget is negative", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse the budget as unusable", func(t *testing.T) {
			t.Parallel()

			window := model_builders.NewPaymentSpendRuleWindowBuilder().WithMaxAmount("-1").Build()

			_, err := window.RecordSpend(big.NewInt(1), now)

			assert.EqualError(t, err, "window budget is not a usable amount")
		})
	})

	t.Run("Given the recorded spend is not a usable amount", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse the recorded spend as unusable", func(t *testing.T) {
			t.Parallel()

			window := model_builders.NewPaymentSpendRuleWindowBuilder().
				WithSpent("not-a-number").
				WithStartedAt(now.Add(-time.Hour)).
				Build()

			_, err := window.RecordSpend(big.NewInt(1), now)

			assert.EqualError(t, err, "window spend is not a usable amount")
		})
	})

	t.Run("Given the recorded spend is negative", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse the recorded spend as unusable", func(t *testing.T) {
			t.Parallel()

			window := model_builders.NewPaymentSpendRuleWindowBuilder().
				WithSpent("-1").
				WithStartedAt(now.Add(-time.Hour)).
				Build()

			_, err := window.RecordSpend(big.NewInt(1), now)

			assert.EqualError(t, err, "window spend is not a usable amount")
		})
	})

	t.Run("Given a period that has never been used", func(t *testing.T) {
		t.Parallel()

		t.Run("Should start the period at the current whole second", func(t *testing.T) {
			t.Parallel()

			window := model_builders.NewPaymentSpendRuleWindowBuilder().Build()

			_, err := window.RecordSpend(big.NewInt(300), now)

			assert.NoError(t, err)
			if assert.NotNil(t, window.StartedAt) {
				assert.Equal(t, now.Truncate(time.Second), *window.StartedAt)
			}
		})

		t.Run("Should record the amount", func(t *testing.T) {
			t.Parallel()

			window := model_builders.NewPaymentSpendRuleWindowBuilder().Build()

			_, err := window.RecordSpend(big.NewInt(300), now)

			assert.NoError(t, err)
			assert.Equal(t, "300", window.Spent)
		})
	})

	t.Run("Given a running period with room left", func(t *testing.T) {
		t.Parallel()

		t.Run("Should allow the payment", func(t *testing.T) {
			t.Parallel()

			window := model_builders.NewPaymentSpendRuleWindowBuilder().
				WithMaxAmount("1000").
				WithSpent("600").
				WithStartedAt(now.Add(-time.Hour)).
				Build()

			recorded, err := window.RecordSpend(big.NewInt(400), now)

			assert.NoError(t, err)
			assert.True(t, recorded)
		})

		t.Run("Should add the amount to what the period has absorbed", func(t *testing.T) {
			t.Parallel()

			window := model_builders.NewPaymentSpendRuleWindowBuilder().
				WithMaxAmount("1000").
				WithSpent("600").
				WithStartedAt(now.Add(-time.Hour)).
				Build()

			_, err := window.RecordSpend(big.NewInt(400), now)

			assert.NoError(t, err)
			assert.Equal(t, "1000", window.Spent)
		})
	})

	t.Run("Given the payment would exceed the period", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse the payment", func(t *testing.T) {
			t.Parallel()

			window := model_builders.NewPaymentSpendRuleWindowBuilder().
				WithMaxAmount("1000").
				WithSpent("900").
				WithStartedAt(now.Add(-time.Hour)).
				Build()

			recorded, err := window.RecordSpend(big.NewInt(200), now)

			assert.NoError(t, err)
			assert.False(t, recorded)
		})

		t.Run("Should leave the recorded spend untouched", func(t *testing.T) {
			t.Parallel()

			window := model_builders.NewPaymentSpendRuleWindowBuilder().
				WithMaxAmount("1000").
				WithSpent("900").
				WithStartedAt(now.Add(-time.Hour)).
				Build()

			_, err := window.RecordSpend(big.NewInt(200), now)

			assert.NoError(t, err)
			assert.Equal(t, "900", window.Spent)
		})
	})

	t.Run("Given the period has elapsed", func(t *testing.T) {
		t.Parallel()

		// The budget returns on its own; nothing needs resetting by hand.
		t.Run("Should measure the payment against a fresh period", func(t *testing.T) {
			t.Parallel()

			window := model_builders.NewPaymentSpendRuleWindowBuilder().
				WithMaxAmount("1000").
				WithSpent("1000").
				WithStartedAt(now.Add(-25 * time.Hour)).
				Build()

			_, err := window.RecordSpend(big.NewInt(1000), now)

			assert.NoError(t, err)
			assert.Equal(t, "1000", window.Spent)
		})

		t.Run("Should restart the period at the current whole second", func(t *testing.T) {
			t.Parallel()

			window := model_builders.NewPaymentSpendRuleWindowBuilder().
				WithMaxAmount("1000").
				WithSpent("1000").
				WithStartedAt(now.Add(-25 * time.Hour)).
				Build()

			_, err := window.RecordSpend(big.NewInt(1000), now)

			assert.NoError(t, err)
			if assert.NotNil(t, window.StartedAt) {
				assert.Equal(t, now.Truncate(time.Second), *window.StartedAt)
			}
		})
	})
}
