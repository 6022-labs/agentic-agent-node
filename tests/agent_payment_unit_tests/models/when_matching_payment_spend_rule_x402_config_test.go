package models_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
)

const (
	matchNetwork  = "eip155:80002"
	matchAsset    = "0x41E94Eb019C0762f9Bfcf9Fb1E58725BfB0e7582"
	matchPayTo    = "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
	matchTransfer = "eip3009"
)

func TestWhenMatchingPaymentSpendRuleX402Config(t *testing.T) {
	t.Parallel()

	t.Run("Given the seeded catch-all config", func(t *testing.T) {
		t.Parallel()

		t.Run("Should match any quote", func(t *testing.T) {
			t.Parallel()

			config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().Build()

			assert.True(t, config.Matches(matchNetwork, matchAsset, matchPayTo, matchTransfer))
		})
	})

	t.Run("Given one facet is narrowed", func(t *testing.T) {
		t.Parallel()

		// Every facet must hit, so narrowing any one of them is enough to refuse.
		narrowed := []struct {
			name   string
			config *models.PaymentSpendRuleX402Config
		}{
			{"network", model_builders.NewPaymentSpendRuleX402ConfigBuilder().
				WithNetworkPatterns([]string{"eip155:137"}).Build()},
			{"asset", model_builders.NewPaymentSpendRuleX402ConfigBuilder().
				WithAssetPatterns([]string{"0x2222222222222222222222222222222222222222"}).Build()},
			{"payTo", model_builders.NewPaymentSpendRuleX402ConfigBuilder().
				WithPayToPatterns([]string{"0x3333333333333333333333333333333333333333"}).Build()},
			{"assetTransferMethod", model_builders.NewPaymentSpendRuleX402ConfigBuilder().
				WithAssetTransferMethodPatterns([]string{"permit2"}).Build()},
		}

		for _, narrowedCase := range narrowed {
			t.Run("Should refuse a quote outside the "+narrowedCase.name+" pattern", func(t *testing.T) {
				t.Parallel()

				assert.False(t, narrowedCase.config.Matches(matchNetwork, matchAsset, matchPayTo, matchTransfer))
			})
		}
	})

	t.Run("Given patterns and values differ only in case", func(t *testing.T) {
		t.Parallel()

		t.Run("Should still match, since addresses vary in checksum casing", func(t *testing.T) {
			t.Parallel()

			config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().
				WithAssetPatterns([]string{"0X41E94EB019C0762F9BFCF9FB1E58725BFB0E7582"}).
				Build()

			assert.True(t, config.Matches(matchNetwork, matchAsset, matchPayTo, matchTransfer))
		})
	})

	t.Run("Given an exact address pattern", func(t *testing.T) {
		t.Parallel()

		t.Run("Should anchor, so a longer address is not a prefix match", func(t *testing.T) {
			t.Parallel()

			config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().
				WithPayToPatterns([]string{"0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"}).
				Build()

			assert.False(t, config.Matches(matchNetwork, matchAsset, matchPayTo+"00", matchTransfer))
		})
	})

	t.Run("Given a pattern whose meaning depends on escape case", func(t *testing.T) {
		t.Parallel()

		// Lowercasing the pattern would rewrite \D into \d and make a rule
		// written to exclude digits match them — a fail-open inversion.
		t.Run("Should not match a negated class the pattern excludes", func(t *testing.T) {
			t.Parallel()

			config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().
				WithNetworkPatterns([]string{`eip155:\D+`}).
				Build()

			assert.False(t, config.Matches(matchNetwork, matchAsset, matchPayTo, matchTransfer))
		})

		t.Run("Should still match the positive class", func(t *testing.T) {
			t.Parallel()

			config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().
				WithNetworkPatterns([]string{`eip155:\d+`}).
				Build()

			assert.True(t, config.Matches(matchNetwork, matchAsset, matchPayTo, matchTransfer))
		})
	})

	t.Run("Given an empty pattern array", func(t *testing.T) {
		t.Parallel()

		t.Run("Should match nothing, so a rule can never widen by omission", func(t *testing.T) {
			t.Parallel()

			config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().
				WithNetworkPatterns([]string{}).
				Build()

			assert.False(t, config.Matches(matchNetwork, matchAsset, matchPayTo, matchTransfer))
		})
	})

	t.Run("Given an uncompilable pattern", func(t *testing.T) {
		t.Parallel()

		t.Run("Should skip it rather than match everything", func(t *testing.T) {
			t.Parallel()

			config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().
				WithAssetPatterns([]string{"0x[("}).
				Build()

			assert.False(t, config.Matches(matchNetwork, matchAsset, matchPayTo, matchTransfer))
		})

		t.Run("Should still honour a valid sibling pattern", func(t *testing.T) {
			t.Parallel()

			config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().
				WithAssetPatterns([]string{"0x[(", matchAsset}).
				Build()

			assert.True(t, config.Matches(matchNetwork, matchAsset, matchPayTo, matchTransfer))
		})
	})
}
