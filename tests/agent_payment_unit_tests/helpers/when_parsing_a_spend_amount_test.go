package helpers_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/helpers"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/stretchr/testify/assert"
)

func TestWhenParsingASpendAmount(t *testing.T) {
	t.Parallel()

	t.Run("Given an amount that is not a whole number", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError on the given field", func(t *testing.T) {
			t.Parallel()

			amount, err := helpers.ParseSpendAmount("field", "1.5")

			assert.Nil(t, amount)
			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "field", validationErr.Field)
		})
	})

	t.Run("Given a negative amount", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError on the given field", func(t *testing.T) {
			t.Parallel()

			amount, err := helpers.ParseSpendAmount("field", "-1")

			assert.Nil(t, amount)
			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "field", validationErr.Field)
		})
	})

	t.Run("Given an amount above the largest uint256", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError on the given field", func(t *testing.T) {
			t.Parallel()

			amount, err := helpers.ParseSpendAmount(
				"field",
				"115792089237316195423570985008687907853269984665640564039457584007913129639936",
			)

			assert.Nil(t, amount)
			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "field", validationErr.Field)
		})
	})

	t.Run("Given a padded whole amount", func(t *testing.T) {
		t.Parallel()

		t.Run("Should parse it", func(t *testing.T) {
			t.Parallel()

			amount, err := helpers.ParseSpendAmount("field", " 007 ")

			assert.NoError(t, err)
			assert.Equal(t, "7", amount.Dec())
		})
	})
}
