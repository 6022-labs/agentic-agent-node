package models_test

import (
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
)

func TestWhenResettingAPaymentSpendWindow(t *testing.T) {
	t.Parallel()

	t.Run("Given a running period with spend recorded", func(t *testing.T) {
		t.Parallel()

		t.Run("Should drop the recorded spend", func(t *testing.T) {
			t.Parallel()

			window := model_builders.NewPaymentSpendRuleWindowBuilder().
				WithSpent("4000").
				WithStartedAt(time.Now().Add(-time.Hour)).
				Build()

			window.Reset()

			assert.Equal(t, "0", window.Spent)
		})

		t.Run("Should leave no period running", func(t *testing.T) {
			t.Parallel()

			window := model_builders.NewPaymentSpendRuleWindowBuilder().
				WithSpent("4000").
				WithStartedAt(time.Now().Add(-time.Hour)).
				Build()

			window.Reset()

			assert.Nil(t, window.StartedAt)
		})
	})
}
