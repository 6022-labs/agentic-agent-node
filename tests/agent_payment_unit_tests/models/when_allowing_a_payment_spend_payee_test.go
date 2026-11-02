package models_test

import (
	"strings"
	"testing"

	common_tests "github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
)

func TestWhenAllowingAPaymentSpendPayee(t *testing.T) {
	t.Parallel()

	t.Run("Given the policy lists no payee", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse every payee", func(t *testing.T) {
			t.Parallel()

			policy := model_builders.NewPaymentSpendPolicyBuilder().WithPayToPatterns([]string{}).Build()

			assert.False(t, policy.AllowsPayee(common_tests.RandomEthAddress().Hex()))
		})
	})

	t.Run("Given the policy lists the payee in another case", func(t *testing.T) {
		t.Parallel()

		t.Run("Should allow the payee", func(t *testing.T) {
			t.Parallel()

			payTo := common_tests.RandomEthAddress().Hex()
			policy := model_builders.NewPaymentSpendPolicyBuilder().
				WithPayToPatterns([]string{strings.ToLower(payTo)}).
				Build()

			assert.True(t, policy.AllowsPayee("0x"+strings.ToUpper(payTo[2:])))
		})
	})

	t.Run("Given the policy lists other payees only", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse the payee", func(t *testing.T) {
			t.Parallel()

			policy := model_builders.NewPaymentSpendPolicyBuilder().
				WithPayToPatterns([]string{common_tests.RandomEthAddress().Hex()}).
				Build()

			assert.False(t, policy.AllowsPayee(common_tests.RandomEthAddress().Hex()))
		})
	})

	// Patterns are full-matched, so a prefix alone never matches.
	t.Run("Given a payee pattern that is only a prefix", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse the payee", func(t *testing.T) {
			t.Parallel()

			payTo := common_tests.RandomEthAddress().Hex()
			policy := model_builders.NewPaymentSpendPolicyBuilder().WithPayToPatterns([]string{payTo[:6]}).Build()

			assert.False(t, policy.AllowsPayee(payTo))
		})
	})

	t.Run("Given a pattern that does not compile", func(t *testing.T) {
		t.Parallel()

		t.Run("Should skip it and keep matching the others", func(t *testing.T) {
			t.Parallel()

			payTo := common_tests.RandomEthAddress().Hex()
			policy := model_builders.NewPaymentSpendPolicyBuilder().WithPayToPatterns([]string{"(", payTo}).Build()

			assert.True(t, policy.AllowsPayee(payTo))
		})
	})
}
