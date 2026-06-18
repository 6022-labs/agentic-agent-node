package helper_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/helper"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
)

// A fixed secp256k1 test key and its derived address.
const testPrivateKeyHex = "4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318"

func expectedFromAddress(t *testing.T) string {
	t.Helper()
	key, err := crypto.HexToECDSA(testPrivateKeyHex)
	assert.NoError(t, err)
	return crypto.PubkeyToAddress(key.PublicKey).Hex()
}

func TestWhenBuildingTransactor(t *testing.T) {
	t.Parallel()

	t.Run("Given a valid private key", func(t *testing.T) {
		t.Parallel()

		t.Run("Should build a transactor bound to the chain and signer", func(t *testing.T) {
			t.Parallel()

			opts, err := helper.Transactor(testPrivateKeyHex, 80002)

			assert.NoError(t, err)
			assert.NotNil(t, opts)
			assert.Equal(t, expectedFromAddress(t), opts.From.Hex())
		})

		t.Run("Should accept a 0x-prefixed, whitespace-padded key", func(t *testing.T) {
			t.Parallel()

			opts, err := helper.Transactor("  0x"+testPrivateKeyHex+"  ", 1)

			assert.NoError(t, err)
			assert.Equal(t, expectedFromAddress(t), opts.From.Hex())
		})
	})

	t.Run("Given an invalid private key", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return an error", func(t *testing.T) {
			t.Parallel()

			_, err := helper.Transactor("not-a-key", 1)

			assert.Error(t, err)
		})
	})
}
