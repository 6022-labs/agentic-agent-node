package services_test

import (
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/gateway/services"
	"github.com/stretchr/testify/assert"
)

func TestWhenGeneratingNonce(t *testing.T) {
	t.Parallel()

	t.Run("Given a fresh nonce store", func(t *testing.T) {
		t.Parallel()

		store := services.NewNonceStore(5 * time.Minute)

		t.Run("Should return a non-empty hex nonce with a future expiry", func(t *testing.T) {
			nonce, expiresAt, err := store.Generate()

			assert.NoError(t, err)
			assert.NotEmpty(t, nonce)
			assert.Len(t, nonce, 32) // 16 bytes hex-encoded
			assert.True(t, expiresAt.After(time.Now()))
		})

		t.Run("Should never return the same nonce twice", func(t *testing.T) {
			a, _, _ := store.Generate()
			b, _, _ := store.Generate()
			assert.NotEqual(t, a, b)
		})
	})

	t.Run("Given a nonce was just generated", func(t *testing.T) {
		t.Parallel()

		store := services.NewNonceStore(5 * time.Minute)
		nonce, _, _ := store.Generate()

		t.Run("Should accept the nonce on first consume", func(t *testing.T) {
			assert.True(t, store.Consume(nonce))
		})

		t.Run("Should reject the nonce on second consume", func(t *testing.T) {
			store2 := services.NewNonceStore(5 * time.Minute)
			n, _, _ := store2.Generate()
			assert.True(t, store2.Consume(n))
			assert.False(t, store2.Consume(n))
		})
	})

	t.Run("Given a nonce that does not exist", func(t *testing.T) {
		t.Parallel()

		store := services.NewNonceStore(5 * time.Minute)

		t.Run("Should return false on consume", func(t *testing.T) {
			assert.False(t, store.Consume("does-not-exist"))
		})
	})

	t.Run("Given a nonce whose TTL has elapsed", func(t *testing.T) {
		t.Parallel()

		store := services.NewNonceStore(time.Nanosecond)
		nonce, _, _ := store.Generate()
		time.Sleep(2 * time.Millisecond)

		t.Run("Should reject the nonce as expired", func(t *testing.T) {
			assert.False(t, store.Consume(nonce))
		})
	})
}
