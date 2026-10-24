package services_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/stretchr/testify/assert"
)

func TestWhenBuildingMintAddresses(t *testing.T) {
	t.Parallel()

	t.Run("Given a blank value for a type", func(t *testing.T) {
		t.Parallel()

		t.Run("Should drop the empty entry", func(t *testing.T) {
			t.Parallel()

			result := identity_interfaces.MintAddress{}.FromMap(map[enums.AddressType]string{
				enums.AddressTypeEvm: "0xabc",
				enums.AddressTypeBtc: "",
			})

			assert.Equal(t, []identity_interfaces.MintAddress{
				{AddressType: enums.AddressTypeEvm, Value: "0xabc"},
			}, result)
		})
	})

	t.Run("Given a nil map", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return an empty, non-nil slice", func(t *testing.T) {
			t.Parallel()

			result := identity_interfaces.MintAddress{}.FromMap(nil)

			assert.NotNil(t, result)
			assert.Empty(t, result)
		})
	})

	t.Run("Given evm and btc addresses", func(t *testing.T) {
		t.Parallel()

		t.Run("Should order evm first regardless of map iteration", func(t *testing.T) {
			t.Parallel()

			result := identity_interfaces.MintAddress{}.FromMap(map[enums.AddressType]string{
				enums.AddressTypeBtc: "bc1qxyz",
				enums.AddressTypeEvm: "0xabc",
			})

			assert.Equal(t, []identity_interfaces.MintAddress{
				{AddressType: enums.AddressTypeEvm, Value: "0xabc"},
				{AddressType: enums.AddressTypeBtc, Value: "bc1qxyz"},
			}, result)
		})
	})
}
