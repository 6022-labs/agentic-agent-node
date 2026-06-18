package services_test

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway/core/services"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
)

type WhenRecoveringSignerTestingSuite struct {
	sut *services.SignatureVerifier
}

func WhenRecoveringSignerBeforeEach() *WhenRecoveringSignerTestingSuite {
	return &WhenRecoveringSignerTestingSuite{
		sut: services.NewSignatureVerifier(),
	}
}

func TestWhenRecoveringSigner(t *testing.T) {
	t.Parallel()

	t.Run("Given a valid EIP-191 signature over a message", func(t *testing.T) {
		t.Parallel()

		t.Run("Should recover the signer address", func(t *testing.T) {
			t.Parallel()

			suite := WhenRecoveringSignerBeforeEach()

			key, _ := crypto.GenerateKey()
			expectedAddress := crypto.PubkeyToAddress(key.PublicKey)
			message := "challenge-nonce-123"

			prefixed := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)
			hash := crypto.Keccak256Hash([]byte(prefixed))
			sig, err := crypto.Sign(hash.Bytes(), key)
			assert.NoError(t, err)
			sig[64] += 27 // many wallets ship v as 27/28
			sigHex := "0x" + hex.EncodeToString(sig)

			recovered, err := suite.sut.Recover(message, sigHex)

			assert.NoError(t, err)
			assert.Equal(t, expectedAddress, recovered)
		})
	})

	t.Run("Given a signature whose v is already 0/1", func(t *testing.T) {
		t.Parallel()

		t.Run("Should still recover the signer", func(t *testing.T) {
			t.Parallel()

			suite := WhenRecoveringSignerBeforeEach()

			key, _ := crypto.GenerateKey()
			expectedAddress := crypto.PubkeyToAddress(key.PublicKey)
			message := "another-message"

			prefixed := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)
			hash := crypto.Keccak256Hash([]byte(prefixed))
			sig, _ := crypto.Sign(hash.Bytes(), key)
			// leave v as 0/1
			sigHex := "0x" + hex.EncodeToString(sig)

			recovered, err := suite.sut.Recover(message, sigHex)

			assert.NoError(t, err)
			assert.Equal(t, expectedAddress, recovered)
		})
	})

	t.Run("Given a signature recovered against a different message", func(t *testing.T) {
		t.Parallel()

		t.Run("Should recover a different address", func(t *testing.T) {
			t.Parallel()

			suite := WhenRecoveringSignerBeforeEach()

			key, _ := crypto.GenerateKey()
			expectedAddress := crypto.PubkeyToAddress(key.PublicKey)
			original := "the-original-message"

			prefixed := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(original), original)
			hash := crypto.Keccak256Hash([]byte(prefixed))
			sig, _ := crypto.Sign(hash.Bytes(), key)
			sigHex := "0x" + hex.EncodeToString(sig)

			recovered, err := suite.sut.Recover("a-different-message", sigHex)

			assert.NoError(t, err)
			assert.NotEqual(t, expectedAddress, recovered)
		})
	})

	t.Run("Given a malformed signature", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject non-hex input", func(t *testing.T) {
			t.Parallel()

			suite := WhenRecoveringSignerBeforeEach()

			_, err := suite.sut.Recover("msg", "0xnothex")
			assert.Error(t, err)
		})

		t.Run("Should reject wrong-length signature", func(t *testing.T) {
			t.Parallel()

			suite := WhenRecoveringSignerBeforeEach()

			_, err := suite.sut.Recover("msg", "0x"+hex.EncodeToString(make([]byte, 10)))
			assert.Error(t, err)
		})

		t.Run("Should reject invalid recovery id", func(t *testing.T) {
			t.Parallel()

			suite := WhenRecoveringSignerBeforeEach()

			bad := make([]byte, 65)
			bad[64] = 99
			_, err := suite.sut.Recover("msg", "0x"+hex.EncodeToString(bad))
			assert.Error(t, err)
		})
	})
}
