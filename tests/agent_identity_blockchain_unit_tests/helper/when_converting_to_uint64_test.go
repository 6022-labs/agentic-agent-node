package helper_test

import (
	"math/big"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/helper"
	"github.com/stretchr/testify/assert"
)

func TestWhenConvertingToUint64(t *testing.T) {
	t.Parallel()

	t.Run("Given a nil value", func(t *testing.T) {
		t.Parallel()

		t.Run("Should collapse to zero", func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, uint64(0), helper.ToUint64(nil))
		})
	})

	t.Run("Given a value overflowing uint64", func(t *testing.T) {
		t.Parallel()

		t.Run("Should collapse to zero", func(t *testing.T) {
			t.Parallel()

			overflow := new(big.Int).Lsh(big.NewInt(1), 80)
			assert.Equal(t, uint64(0), helper.ToUint64(overflow))
		})
	})

	t.Run("Given a value within uint64 range", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return it as uint64", func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, uint64(42), helper.ToUint64(big.NewInt(42)))
		})
	})
}
