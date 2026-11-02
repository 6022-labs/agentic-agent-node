package models_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
)

func TestWhenAcceptingAnX402SpendAssetTransferMethod(t *testing.T) {
	t.Parallel()

	t.Run("Given the rule names no transfer method", func(t *testing.T) {
		t.Parallel()

		t.Run("Should accept permit2", func(t *testing.T) {
			t.Parallel()

			config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().Build()

			assert.True(t, config.AcceptsAssetTransferMethod("permit2"))
		})
	})

	t.Run("Given the rule demands eip3009 and the quote spells eip3009 in another case", func(t *testing.T) {
		t.Parallel()

		t.Run("Should accept the quote", func(t *testing.T) {
			t.Parallel()

			config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().
				WithAssetTransferMethod(enums.AssetTransferMethodEip3009).
				Build()

			assert.True(t, config.AcceptsAssetTransferMethod(" EIP3009 "))
		})
	})

	// permit2 grants a standing max-uint256 allowance, so the constraint has teeth.
	t.Run("Given the rule demands eip3009 and the quote uses permit2", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse the quote", func(t *testing.T) {
			t.Parallel()

			config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().
				WithAssetTransferMethod(enums.AssetTransferMethodEip3009).
				Build()

			assert.False(t, config.AcceptsAssetTransferMethod("permit2"))
		})
	})
}
