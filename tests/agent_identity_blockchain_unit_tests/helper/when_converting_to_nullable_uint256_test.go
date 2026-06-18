package helper_test

import (
	"math/big"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/helper"
	"github.com/stretchr/testify/assert"
)

func TestWhenConvertingToNullableUint256(t *testing.T) {
	t.Parallel()

	t.Run("Given a nil value", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return an unset NullableUint256", func(t *testing.T) {
			t.Parallel()

			result := helper.ToNullableUint256(nil)

			assert.False(t, result.IsSet)
			assert.Equal(t, big.NewInt(0), result.Value)
		})
	})

	t.Run("Given a value", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a set NullableUint256 carrying that value", func(t *testing.T) {
			t.Parallel()

			cloneOf := uint64(42)
			result := helper.ToNullableUint256(&cloneOf)

			assert.True(t, result.IsSet)
			assert.Equal(t, big.NewInt(42), result.Value)
		})
	})
}
