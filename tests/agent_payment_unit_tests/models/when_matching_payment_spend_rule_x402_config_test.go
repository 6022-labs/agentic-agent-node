package models_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
)

const (
	quoteNetwork = "eip155:80002"
	quoteAsset   = "0x41E94Eb019C0762f9Bfcf9Fb1E58725BfB0e7582"
	quotePayTo   = "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
)

func TestWhenMatchingPaymentSpendRuleX402Config(t *testing.T) {
	t.Parallel()

	t.Run("Given the quote is the money the rule describes", func(t *testing.T) {
		t.Parallel()

		t.Run("Should match whatever case the addresses arrive in", func(t *testing.T) {
			t.Parallel()

			config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().Build()

			assert.True(t, config.Matches(quoteNetwork, quoteAsset, quotePayTo, "eip3009"))
		})
	})

	t.Run("Given the quote is the same token on another chain", func(t *testing.T) {
		t.Parallel()

		// An ERC-20 address is only unique per network, so the network is compared, never patterned.
		t.Run("Should not match", func(t *testing.T) {
			t.Parallel()

			config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().Build()

			assert.False(t, config.Matches("eip155:1", quoteAsset, quotePayTo, "eip3009"))
		})
	})

	t.Run("Given the quote is a different token", func(t *testing.T) {
		t.Parallel()

		t.Run("Should not match", func(t *testing.T) {
			t.Parallel()

			config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().Build()

			assert.False(t, config.Matches(
				quoteNetwork,
				"0x0000000000000000000000000000000000000001",
				quotePayTo,
				"eip3009",
			))
		})
	})

	t.Run("Given the payee is outside the rule's patterns", func(t *testing.T) {
		t.Parallel()

		t.Run("Should not match", func(t *testing.T) {
			t.Parallel()

			config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().
				WithPayToPatterns([]string{"0x0000000000000000000000000000000000000000"}).
				Build()

			assert.False(t, config.Matches(quoteNetwork, quoteAsset, quotePayTo, "eip3009"))
		})
	})

	t.Run("Given the rule demands one-shot eip3009", func(t *testing.T) {
		t.Parallel()

		// permit2 grants a standing max-uint256 allowance, so the constraint has teeth.
		t.Run("Should refuse a permit2 quote", func(t *testing.T) {
			t.Parallel()

			config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().
				WithAssetTransferMethod(enums.AssetTransferMethodEip3009).
				Build()

			assert.True(t, config.Matches(quoteNetwork, quoteAsset, quotePayTo, "eip3009"))
			assert.False(t, config.Matches(quoteNetwork, quoteAsset, quotePayTo, "permit2"))
		})
	})

	t.Run("Given the rule names no transfer method", func(t *testing.T) {
		t.Parallel()

		t.Run("Should accept whichever the token uses", func(t *testing.T) {
			t.Parallel()

			config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().Build()

			assert.True(t, config.Matches(quoteNetwork, quoteAsset, quotePayTo, "eip3009"))
			assert.True(t, config.Matches(quoteNetwork, quoteAsset, quotePayTo, "permit2"))
		})
	})

	t.Run("Given a payee pattern that is only a prefix", func(t *testing.T) {
		t.Parallel()

		t.Run("Should not match, since patterns are full-matched", func(t *testing.T) {
			t.Parallel()

			config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().
				WithPayToPatterns([]string{"0xf39f"}).
				Build()

			assert.False(t, config.Matches(quoteNetwork, quoteAsset, quotePayTo, "eip3009"))
		})
	})
}
