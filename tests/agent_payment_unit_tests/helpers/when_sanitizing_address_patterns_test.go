package helpers_test

import (
	"strings"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/helpers"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/stretchr/testify/assert"
)

func TestWhenSanitizingAddressPatterns(t *testing.T) {
	t.Parallel()

	t.Run("Given a blank pattern", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError on the given field", func(t *testing.T) {
			t.Parallel()

			patterns, err := helpers.SanitizeAddressPatterns("field", []string{"0xabc.*", "  "})

			assert.Nil(t, patterns)
			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "field", validationErr.Field)
		})
	})

	t.Run("Given a pattern that is not a valid regex", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError on the given field", func(t *testing.T) {
			t.Parallel()

			patterns, err := helpers.SanitizeAddressPatterns("field", []string{"0x[("})

			assert.Nil(t, patterns)
			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "field", validationErr.Field)
		})
	})

	t.Run("Given no pattern", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return an empty list", func(t *testing.T) {
			t.Parallel()

			patterns, err := helpers.SanitizeAddressPatterns("field", []string{})

			assert.NoError(t, err)
			assert.Empty(t, patterns)
		})
	})

	t.Run("Given a pattern with surrounding spaces", func(t *testing.T) {
		t.Parallel()

		t.Run("Should trim it", func(t *testing.T) {
			t.Parallel()

			patterns, err := helpers.SanitizeAddressPatterns("field", []string{"  0xabc.*  "})

			assert.NoError(t, err)
			assert.Equal(t, []string{"0xabc.*"}, patterns)
		})
	})

	t.Run("Given a checksummed address", func(t *testing.T) {
		t.Parallel()

		t.Run("Should store it lowercased", func(t *testing.T) {
			t.Parallel()

			address := common_tests.RandomEthAddress().Hex()

			patterns, err := helpers.SanitizeAddressPatterns("field", []string{address})

			assert.NoError(t, err)
			assert.Equal(t, []string{strings.ToLower(address)}, patterns)
		})
	})

	t.Run("Given the same address in two cases", func(t *testing.T) {
		t.Parallel()

		t.Run("Should keep it once", func(t *testing.T) {
			t.Parallel()

			address := common_tests.RandomEthAddress().Hex()

			patterns, err := helpers.SanitizeAddressPatterns("field", []string{address, strings.ToLower(address)})

			assert.NoError(t, err)
			assert.Equal(t, []string{strings.ToLower(address)}, patterns)
		})
	})

	t.Run("Given anyone among other patterns", func(t *testing.T) {
		t.Parallel()

		t.Run("Should keep only anyone", func(t *testing.T) {
			t.Parallel()

			patterns, err := helpers.SanitizeAddressPatterns(
				"field",
				[]string{strings.ToLower(common_tests.RandomEthAddress().Hex()), ".*"},
			)

			assert.NoError(t, err)
			assert.Equal(t, []string{".*"}, patterns)
		})
	})
}
