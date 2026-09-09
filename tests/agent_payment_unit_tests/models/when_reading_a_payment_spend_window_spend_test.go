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

			rule := model_builders.NewPaymentSpendRuleBuilder().
				WithWindowSeconds(86400).
				WithMaxAmountPerWindow("50000").
				WithWindowSpent("4000").
				WithWindowStartedAt(time.Now().Add(-time.Hour)).
				Build()

			assert.Equal(t, "4000", rule.SpentInWindow(time.Now()))
		})
	})

	t.Run("Given the period has elapsed without rolling over", func(t *testing.T) {
		t.Parallel()

		// The row keeps the old spend until the next payment rolls it, so a raw read would overstate it.
		t.Run("Should report nothing spent", func(t *testing.T) {
			t.Parallel()

			rule := model_builders.NewPaymentSpendRuleBuilder().
				WithWindowSeconds(86400).
				WithMaxAmountPerWindow("50000").
				WithWindowSpent("4000").
				WithWindowStartedAt(time.Now().Add(-48 * time.Hour)).
				Build()

			assert.Equal(t, "0", rule.SpentInWindow(time.Now()))
		})
	})

	t.Run("Given a period that has never been used", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report nothing spent", func(t *testing.T) {
			t.Parallel()

			rule := model_builders.NewPaymentSpendRuleBuilder().
				WithWindowSeconds(86400).
				WithMaxAmountPerWindow("50000").
				Build()

			assert.Equal(t, "0", rule.SpentInWindow(time.Now()))
		})
	})

	t.Run("Given a rule bounded per payment only", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report nothing spent", func(t *testing.T) {
			t.Parallel()

			rule := model_builders.NewPaymentSpendRuleBuilder().Build()

			assert.Equal(t, "0", rule.SpentInWindow(time.Now()))
		})
	})
}
