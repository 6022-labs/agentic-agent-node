package models_test

import (
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
)

func TestWhenReadingAPaymentSpendWindowSpend(t *testing.T) {
	t.Parallel()

	t.Run("Given a running period with spend recorded", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report what the period has absorbed", func(t *testing.T) {
			t.Parallel()

			window := model_builders.NewPaymentSpendRuleWindowBuilder().
				WithSpent("4000").
				WithStartedAt(time.Now().Add(-time.Hour)).
				Build()

			assert.Equal(t, "4000", window.SpentAt(time.Now()))
		})
	})

	t.Run("Given the period has elapsed without rolling over", func(t *testing.T) {
		t.Parallel()

		// The row keeps the old spend until the next payment rolls it, so a raw read would overstate it.
		t.Run("Should report nothing spent", func(t *testing.T) {
			t.Parallel()

			window := model_builders.NewPaymentSpendRuleWindowBuilder().
				WithSpent("4000").
				WithStartedAt(time.Now().Add(-48 * time.Hour)).
				Build()

			assert.Equal(t, "0", window.SpentAt(time.Now()))
		})
	})

	// The end instant already belongs to the next period.
	t.Run("Given the period ends exactly now", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report nothing spent", func(t *testing.T) {
			t.Parallel()

			startedAt := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
			window := model_builders.NewPaymentSpendRuleWindowBuilder().
				WithSpent("4000").
				WithStartedAt(startedAt).
				Build()

			assert.Equal(t, "0", window.SpentAt(startedAt.Add(24*time.Hour)))
		})
	})

	t.Run("Given a period that has never been used", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report nothing spent", func(t *testing.T) {
			t.Parallel()

			window := model_builders.NewPaymentSpendRuleWindowBuilder().Build()

			assert.Equal(t, "0", window.SpentAt(time.Now()))
		})
	})
}
