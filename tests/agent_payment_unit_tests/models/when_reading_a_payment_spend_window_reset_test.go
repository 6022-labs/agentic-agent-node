package models_test

import (
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
)

func TestWhenReadingAPaymentSpendWindowReset(t *testing.T) {
	t.Parallel()

	t.Run("Given a running period", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report when it rolls over", func(t *testing.T) {
			t.Parallel()

			startedAt := time.Now().Add(-time.Hour)
			rule := model_builders.NewPaymentSpendRuleBuilder().
				WithWindowSeconds(86400).
				WithMaxAmountPerWindow("50000").
				WithWindowSpent("4000").
				WithWindowStartedAt(startedAt).
				Build()

			resetsAt := rule.WindowResetsAt(time.Now())

			if assert.NotNil(t, resetsAt) {
				assert.Equal(t, startedAt.Add(24*time.Hour).Unix(), resetsAt.Unix())
			}
		})
	})

	t.Run("Given the period has elapsed without rolling over", func(t *testing.T) {
		t.Parallel()

		// The next payment restarts it, so there is no pending reset to announce.
		t.Run("Should report no reset pending", func(t *testing.T) {
			t.Parallel()

			rule := model_builders.NewPaymentSpendRuleBuilder().
				WithWindowSeconds(86400).
				WithMaxAmountPerWindow("50000").
				WithWindowSpent("4000").
				WithWindowStartedAt(time.Now().Add(-48 * time.Hour)).
				Build()

			assert.Nil(t, rule.WindowResetsAt(time.Now()))
		})
	})

	t.Run("Given a period that has never been used", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report no reset pending", func(t *testing.T) {
			t.Parallel()

			rule := model_builders.NewPaymentSpendRuleBuilder().
				WithWindowSeconds(86400).
				WithMaxAmountPerWindow("50000").
				Build()

			assert.Nil(t, rule.WindowResetsAt(time.Now()))
		})
	})

	t.Run("Given a rule bounded per payment only", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report no reset pending", func(t *testing.T) {
			t.Parallel()

			rule := model_builders.NewPaymentSpendRuleBuilder().Build()

			assert.Nil(t, rule.WindowResetsAt(time.Now()))
		})
	})
}
