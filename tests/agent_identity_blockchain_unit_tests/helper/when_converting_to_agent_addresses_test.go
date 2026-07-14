package helper_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/helper"
	"github.com/stretchr/testify/assert"
)

func TestWhenConvertingToAgentAddresses(t *testing.T) {
	t.Parallel()

	t.Run("Given mint addresses", func(t *testing.T) {
		t.Parallel()

		t.Run("Should map each entry preserving type and value", func(t *testing.T) {
			t.Parallel()

			result := helper.ToAgentAddresses([]identity_interfaces.MintAddress{
				{AddressType: enums.AddressTypeEvm, Value: "0xabc"},
				{AddressType: enums.AddressTypeBtc, Value: "bc1qxyz"},
			})

			assert.Len(t, result, 2)
			assert.Equal(t, "evm", result[0].AddressType)
			assert.Equal(t, "0xabc", result[0].Value)
			assert.Equal(t, "btc", result[1].AddressType)
			assert.Equal(t, "bc1qxyz", result[1].Value)
		})
	})

	t.Run("Given no entries", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return an empty, non-nil slice", func(t *testing.T) {
			t.Parallel()

			result := helper.ToAgentAddresses(nil)

			assert.NotNil(t, result)
			assert.Empty(t, result)
		})
	})
}
