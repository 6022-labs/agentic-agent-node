package enums_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/common/x402"
	"github.com/stretchr/testify/assert"
)

// The gateway matches on the x402 package's copies while the runtime matches on
// the enum, so the two must keep agreeing or a payment verifies as neither.
func TestWhenMatchingX402WireValues(t *testing.T) {
	t.Parallel()

	t.Run("Given the x402 wire constants", func(t *testing.T) {
		t.Parallel()

		t.Run("Should equal the asset transfer method enum", func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, x402.AssetTransferMethodEip3009, enums.AssetTransferMethodEip3009.String())
			assert.Equal(t, x402.AssetTransferMethodPermit2, enums.AssetTransferMethodPermit2.String())
		})
	})
}
