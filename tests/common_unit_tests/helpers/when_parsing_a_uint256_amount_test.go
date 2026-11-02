package helpers_test

import (
	"math/big"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/common/helpers"
	"github.com/stretchr/testify/assert"
)

const maxUint256 = "115792089237316195423570985008687907853269984665640564039457584007913129639935"

func TestWhenParsingAUint256Amount(t *testing.T) {
	t.Parallel()

	t.Run("Given a negative amount", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse it", func(t *testing.T) {
			t.Parallel()

			_, ok := helpers.ParseUint256("-1")

			assert.False(t, ok)
		})
	})

	t.Run("Given an amount that is not a base-10 integer", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse it", func(t *testing.T) {
			t.Parallel()

			_, ok := helpers.ParseUint256("1.5")

			assert.False(t, ok)
		})
	})

	t.Run("Given an amount one above the largest uint256", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse it", func(t *testing.T) {
			t.Parallel()

			_, ok := helpers.ParseUint256(
				"115792089237316195423570985008687907853269984665640564039457584007913129639936",
			)

			assert.False(t, ok)
		})
	})

	t.Run("Given the largest uint256", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return it parsed", func(t *testing.T) {
			t.Parallel()

			amount, ok := helpers.ParseUint256(maxUint256)

			assert.True(t, ok)
			expected, _ := new(big.Int).SetString(maxUint256, 10)
			assert.Equal(t, 0, expected.Cmp(amount))
		})
	})

	t.Run("Given zero", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return it parsed", func(t *testing.T) {
			t.Parallel()

			amount, ok := helpers.ParseUint256("0")

			assert.True(t, ok)
			assert.Equal(t, 0, amount.Sign())
		})
	})
}
