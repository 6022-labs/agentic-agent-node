package helper_test

import (
	"strings"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment_blockchain/helper"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
)

// A fixed secp256k1 test key and its derived address.
const transactorPrivateKeyHex = "4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318"

func expectedTransactorAddress(t *testing.T) string {
	t.Helper()
	key, err := crypto.HexToECDSA(transactorPrivateKeyHex)
	assert.NoError(t, err)
	return crypto.PubkeyToAddress(key.PublicKey).Hex()
}

func TestWhenBuildingATransactor(t *testing.T) {
	t.Parallel()

	t.Run("Given an invalid private key", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return an error", func(t *testing.T) {
			t.Parallel()

			_, err := helper.Transactor("not-a-key", 1)

			assert.ErrorContains(t, err, "invalid wallet private key")
		})
	})

	t.Run("Given a 0x-prefixed private key", func(t *testing.T) {
		t.Parallel()

		t.Run("Should sign from the key's address", func(t *testing.T) {
			t.Parallel()

			opts, err := helper.Transactor("0x"+transactorPrivateKeyHex, 80002)

			assert.NoError(t, err)
			assert.Equal(t, expectedTransactorAddress(t), opts.From.Hex())
		})
	})

	t.Run("Given a 0X-prefixed private key", func(t *testing.T) {
		t.Parallel()

		t.Run("Should sign from the key's address", func(t *testing.T) {
			t.Parallel()

			opts, err := helper.Transactor("0X"+strings.ToUpper(transactorPrivateKeyHex), 80002)

			assert.NoError(t, err)
			assert.Equal(t, expectedTransactorAddress(t), opts.From.Hex())
		})
	})

	t.Run("Given a bare hex private key", func(t *testing.T) {
		t.Parallel()

		t.Run("Should sign from the key's address", func(t *testing.T) {
			t.Parallel()

			opts, err := helper.Transactor(transactorPrivateKeyHex, 80002)

			assert.NoError(t, err)
			assert.Equal(t, expectedTransactorAddress(t), opts.From.Hex())
		})
	})
}
