package services_test

import (
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/gateway/core/services"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

var nonceWallet = common.HexToAddress("0xd389d997Df5654b87914492E1e0C6861bfe835DE")

type WhenGeneratingNonceTestingSuite struct {
	sut *services.NonceStore
}

func WhenGeneratingNonceBeforeEach(ttl time.Duration) *WhenGeneratingNonceTestingSuite {
	return &WhenGeneratingNonceTestingSuite{
		sut: services.NewNonceStore(ttl),
	}
}

func TestWhenGeneratingNonce(t *testing.T) {
	t.Parallel()

	t.Run("Given a fresh nonce store", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a challenge with a non-empty hex nonce", func(t *testing.T) {
			t.Parallel()

			suite := WhenGeneratingNonceBeforeEach(5 * time.Minute)

			entry, err := suite.sut.Generate(nonceWallet, 80002)

			assert.NoError(t, err)
			assert.Len(t, entry.Nonce, 64) // 32 bytes hex-encoded
			assert.Equal(t, uint64(80002), entry.ChainId)
			assert.True(t, entry.ExpiresAt.After(time.Now()))
			assert.True(t, entry.ExpiresAt.After(entry.IssuedAt))
		})

		t.Run("Should never return the same nonce twice", func(t *testing.T) {
			t.Parallel()

			suite := WhenGeneratingNonceBeforeEach(5 * time.Minute)

			a, _ := suite.sut.Generate(nonceWallet, 80002)
			b, _ := suite.sut.Generate(nonceWallet, 80002)

			assert.NotEqual(t, a.Nonce, b.Nonce)
		})
	})

	t.Run("Given a challenge was just generated", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return the stored entry on first consume", func(t *testing.T) {
			t.Parallel()

			suite := WhenGeneratingNonceBeforeEach(5 * time.Minute)
			generated, _ := suite.sut.Generate(nonceWallet, 80002)

			entry := suite.sut.Consume(nonceWallet)

			if assert.NotNil(t, entry) {
				assert.Equal(t, *generated, *entry)
			}
		})

		t.Run("Should reject the wallet on second consume", func(t *testing.T) {
			t.Parallel()

			suite := WhenGeneratingNonceBeforeEach(5 * time.Minute)
			_, _ = suite.sut.Generate(nonceWallet, 80002)

			assert.NotNil(t, suite.sut.Consume(nonceWallet))
			assert.Nil(t, suite.sut.Consume(nonceWallet))
		})

		t.Run("Should replace the previous challenge on regenerate", func(t *testing.T) {
			t.Parallel()

			suite := WhenGeneratingNonceBeforeEach(5 * time.Minute)
			_, _ = suite.sut.Generate(nonceWallet, 80002)
			second, _ := suite.sut.Generate(nonceWallet, 80002)

			entry := suite.sut.Consume(nonceWallet)

			if assert.NotNil(t, entry) {
				assert.Equal(t, second.Nonce, entry.Nonce)
			}
		})
	})

	t.Run("Given a wallet without a challenge", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return nil on consume", func(t *testing.T) {
			t.Parallel()

			suite := WhenGeneratingNonceBeforeEach(5 * time.Minute)

			assert.Nil(t, suite.sut.Consume(common.HexToAddress("0x1111111111111111111111111111111111111111")))
		})
	})

	t.Run("Given a challenge whose TTL has elapsed", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject the challenge as expired", func(t *testing.T) {
			t.Parallel()

			suite := WhenGeneratingNonceBeforeEach(time.Nanosecond)
			_, _ = suite.sut.Generate(nonceWallet, 80002)
			time.Sleep(2 * time.Millisecond)

			assert.Nil(t, suite.sut.Consume(nonceWallet))
		})
	})
}
