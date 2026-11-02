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
			window := model_builders.NewPaymentSpendRuleWindowBuilder().
				WithWindowSeconds(86400).
				WithSpent("4000").
				WithStartedAt(startedAt).
				Build()

			resetsAt := window.ResetsAt(time.Now())

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

			window := model_builders.NewPaymentSpendRuleWindowBuilder().
				WithSpent("4000").
				WithStartedAt(time.Now().Add(-48 * time.Hour)).
				Build()

			assert.Nil(t, window.ResetsAt(time.Now()))
		})
	})

	// The end instant already belongs to the next period.
	t.Run("Given the period ends exactly now", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report no reset pending", func(t *testing.T) {
			t.Parallel()

			startedAt := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
			window := model_builders.NewPaymentSpendRuleWindowBuilder().WithStartedAt(startedAt).Build()

			assert.Nil(t, window.ResetsAt(startedAt.Add(24*time.Hour)))
		})
	})

	t.Run("Given a period that has never been used", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report no reset pending", func(t *testing.T) {
			t.Parallel()

			window := model_builders.NewPaymentSpendRuleWindowBuilder().Build()

			assert.Nil(t, window.ResetsAt(time.Now()))
		})
	})
}
