package helper_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment_blockchain/helper"
	"github.com/stretchr/testify/assert"
)

const (
	domainSeparatorChainId = int64(80002)
	domainSeparatorAsset   = "0x41E94Eb019C0762f9Bfcf9Fb1E58725BfB0e7582"
)

func TestWhenComputingDomainSeparator(t *testing.T) {
	t.Parallel()

	t.Run("Given a token EIP-712 domain", func(t *testing.T) {
		t.Parallel()

		t.Run("Should match the separator computed from the EIP-712 spec", func(t *testing.T) {
			t.Parallel()

			separator := helper.ComputeDomainSeparator("USD Coin", "2", domainSeparatorChainId, domainSeparatorAsset)

			expected := expectedDomainSeparator("USD Coin", "2", domainSeparatorChainId, domainSeparatorAsset)
			assert.Equal(t, expected, separator)
		})

		t.Run("Should be deterministic for the same inputs", func(t *testing.T) {
			t.Parallel()

			first := helper.ComputeDomainSeparator("USD Coin", "2", domainSeparatorChainId, domainSeparatorAsset)
			second := helper.ComputeDomainSeparator("USD Coin", "2", domainSeparatorChainId, domainSeparatorAsset)

			assert.Equal(t, first, second)
		})
	})

	t.Run("Given a different domain field", func(t *testing.T) {
		t.Parallel()

		base := helper.ComputeDomainSeparator("USD Coin", "2", domainSeparatorChainId, domainSeparatorAsset)

		t.Run("Should change when the name differs", func(t *testing.T) {
			t.Parallel()

			other := helper.ComputeDomainSeparator("USDC", "2", domainSeparatorChainId, domainSeparatorAsset)
			assert.NotEqual(t, base, other)
		})

		t.Run("Should change when the version differs", func(t *testing.T) {
			t.Parallel()

			other := helper.ComputeDomainSeparator("USD Coin", "1", domainSeparatorChainId, domainSeparatorAsset)
			assert.NotEqual(t, base, other)
		})

		t.Run("Should change when the chain id differs", func(t *testing.T) {
			t.Parallel()

			other := helper.ComputeDomainSeparator("USD Coin", "2", 1, domainSeparatorAsset)
			assert.NotEqual(t, base, other)
		})

		t.Run("Should change when the verifying contract differs", func(t *testing.T) {
			t.Parallel()

			other := helper.ComputeDomainSeparator(
				"USD Coin", "2", domainSeparatorChainId, "0x0000000000000000000000000000000000000001")
			assert.NotEqual(t, base, other)
		})
	})
}
