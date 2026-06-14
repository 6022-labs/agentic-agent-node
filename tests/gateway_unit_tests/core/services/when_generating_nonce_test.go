package services_test

import (
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/gateway/core/services"
	"github.com/stretchr/testify/assert"
)

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

		t.Run("Should return a non-empty hex nonce with a future expiry", func(t *testing.T) {
			t.Parallel()

			suite := WhenGeneratingNonceBeforeEach(5 * time.Minute)

			nonce, expiresAt, err := suite.sut.Generate()

			assert.NoError(t, err)
			assert.NotEmpty(t, nonce)
			assert.Len(t, nonce, 32) // 16 bytes hex-encoded
			assert.True(t, expiresAt.After(time.Now()))
		})

		t.Run("Should never return the same nonce twice", func(t *testing.T) {
			t.Parallel()

			suite := WhenGeneratingNonceBeforeEach(5 * time.Minute)

			a, _, _ := suite.sut.Generate()
			b, _, _ := suite.sut.Generate()

			assert.NotEqual(t, a, b)
		})
	})

	t.Run("Given a nonce was just generated", func(t *testing.T) {
		t.Parallel()

		t.Run("Should accept the nonce on first consume", func(t *testing.T) {
			t.Parallel()

			suite := WhenGeneratingNonceBeforeEach(5 * time.Minute)
			nonce, _, _ := suite.sut.Generate()

			assert.True(t, suite.sut.Consume(nonce))
		})

		t.Run("Should reject the nonce on second consume", func(t *testing.T) {
			t.Parallel()

			suite := WhenGeneratingNonceBeforeEach(5 * time.Minute)
			nonce, _, _ := suite.sut.Generate()

			assert.True(t, suite.sut.Consume(nonce))
			assert.False(t, suite.sut.Consume(nonce))
		})
	})

	t.Run("Given a nonce that does not exist", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return false on consume", func(t *testing.T) {
			t.Parallel()

			suite := WhenGeneratingNonceBeforeEach(5 * time.Minute)

			assert.False(t, suite.sut.Consume("does-not-exist"))
		})
	})

	t.Run("Given a nonce whose TTL has elapsed", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject the nonce as expired", func(t *testing.T) {
			t.Parallel()

			suite := WhenGeneratingNonceBeforeEach(time.Nanosecond)
			nonce, _, _ := suite.sut.Generate()
			time.Sleep(2 * time.Millisecond)

			assert.False(t, suite.sut.Consume(nonce))
		})
	})
}
