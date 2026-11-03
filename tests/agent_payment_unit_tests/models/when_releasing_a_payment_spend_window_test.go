package models_test

import (
	"math/big"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
)

func TestWhenReleasingAPaymentSpendWindow(t *testing.T) {
	t.Parallel()

	periodStartedAt := time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)

	t.Run("Given a period that has never been used", func(t *testing.T) {
		t.Parallel()

		t.Run("Should leave the spend unchanged", func(t *testing.T) {
			t.Parallel()

			window := model_builders.NewPaymentSpendRuleWindowBuilder().Build()

			err := window.Release(big.NewInt(100), periodStartedAt)

			assert.NoError(t, err)
			assert.Equal(t, "0", window.Spent)
		})
	})

	t.Run("Given the period rolled over since the amount was charged", func(t *testing.T) {
		t.Parallel()

		t.Run("Should leave the new period's spend unchanged", func(t *testing.T) {
			t.Parallel()

			window := model_builders.NewPaymentSpendRuleWindowBuilder().
				WithSpent("500").
				WithStartedAt(periodStartedAt.Add(24 * time.Hour)).
				Build()

			err := window.Release(big.NewInt(100), periodStartedAt)

			assert.NoError(t, err)
			assert.Equal(t, "500", window.Spent)
		})
	})

	t.Run("Given the recorded spend is not a usable amount", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse the recorded spend as unusable", func(t *testing.T) {
			t.Parallel()

			window := model_builders.NewPaymentSpendRuleWindowBuilder().
				WithSpent("not-a-number").
				WithStartedAt(periodStartedAt).
				Build()

			err := window.Release(big.NewInt(100), periodStartedAt)

			assert.EqualError(t, err, "window spend is not a usable amount")
		})
	})

	t.Run("Given the amount was charged in the running period", func(t *testing.T) {
		t.Parallel()

		t.Run("Should take the amount back", func(t *testing.T) {
			t.Parallel()

			window := model_builders.NewPaymentSpendRuleWindowBuilder().
				WithSpent("500").
				WithStartedAt(periodStartedAt).
				Build()

			err := window.Release(big.NewInt(100), periodStartedAt)

			assert.NoError(t, err)
			assert.Equal(t, "400", window.Spent)
		})

		t.Run("Should keep the period running", func(t *testing.T) {
			t.Parallel()

			window := model_builders.NewPaymentSpendRuleWindowBuilder().
				WithSpent("500").
				WithStartedAt(periodStartedAt).
				Build()

			err := window.Release(big.NewInt(100), periodStartedAt)

			assert.NoError(t, err)
			if assert.NotNil(t, window.StartedAt) {
				assert.Equal(t, periodStartedAt, *window.StartedAt)
			}
		})
	})

	t.Run("Given the amount exceeds what the period holds", func(t *testing.T) {
		t.Parallel()

		t.Run("Should floor the spend at zero", func(t *testing.T) {
			t.Parallel()

			window := model_builders.NewPaymentSpendRuleWindowBuilder().
				WithSpent("50").
				WithStartedAt(periodStartedAt).
				Build()

			err := window.Release(big.NewInt(100), periodStartedAt)

			assert.NoError(t, err)
			assert.Equal(t, "0", window.Spent)
		})
	})
}
