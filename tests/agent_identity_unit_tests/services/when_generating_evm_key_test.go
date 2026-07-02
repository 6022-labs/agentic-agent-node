package services_test

import (
	"strings"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	identity_services "github.com/6022-labs/agentic-agent-node/src/agent_identity/services"
	"github.com/stretchr/testify/assert"
)

func TestWhenGeneratingEvmKey(t *testing.T) {
	t.Parallel()

	t.Run("Given the evm key generator", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report the evm address type", func(t *testing.T) {
			t.Parallel()

			sut := identity_services.NewEvmKeyGenerator()

			assert.Equal(t, enums.AddressTypeEvm, sut.Type())
		})

		t.Run("Should generate a non-empty keypair with a 0x address", func(t *testing.T) {
			t.Parallel()

			sut := identity_services.NewEvmKeyGenerator()

			cred, err := sut.Generate()

			assert.NoError(t, err)
			assert.True(t, strings.HasPrefix(cred.PrivateKeyHex, "0x"))
			assert.True(t, strings.HasPrefix(cred.Address, "0x"))
			assert.Len(t, cred.Address, 42)
		})

		t.Run("Should generate a distinct keypair on each call", func(t *testing.T) {
			t.Parallel()

			sut := identity_services.NewEvmKeyGenerator()

			first, err := sut.Generate()
			assert.NoError(t, err)
			second, err := sut.Generate()
			assert.NoError(t, err)

			assert.NotEqual(t, first.PrivateKeyHex, second.PrivateKeyHex)
			assert.NotEqual(t, first.Address, second.Address)
		})
	})
}
