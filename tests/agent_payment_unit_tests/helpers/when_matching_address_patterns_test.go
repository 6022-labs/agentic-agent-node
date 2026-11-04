package helpers_test

import (
	"strings"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/helpers"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/stretchr/testify/assert"
)

func TestWhenMatchingAddressPatterns(t *testing.T) {
	t.Parallel()

	// The EIP-55 checksummed form of an address stored lowercased.
	checksummed := common_tests.RandomEthAddress().Hex()

	t.Run("Given the pattern is the lowercase address", func(t *testing.T) {
		t.Parallel()

		t.Run("Should match the caller whatever case they present", func(t *testing.T) {
			t.Parallel()

			assert.True(t, helpers.MatchesAddressPatterns([]string{strings.ToLower(checksummed)}, checksummed))
			assert.True(
				t,
				helpers.MatchesAddressPatterns([]string{strings.ToLower(checksummed)}, strings.ToLower(checksummed)),
			)
		})
	})

	t.Run("Given the pattern spells the address in checksummed case", func(t *testing.T) {
		t.Parallel()

		// Deliberate: folding case would hand free prices to payers who used to pay.
		t.Run("Should not match", func(t *testing.T) {
			t.Parallel()

			assert.False(t, helpers.MatchesAddressPatterns([]string{checksummed}, checksummed))
		})
	})

	t.Run("Given a class that negates uppercase hex", func(t *testing.T) {
		t.Parallel()

		// Folding case would make the negated class exclude a-f too and reject every address.
		t.Run("Should still match every address", func(t *testing.T) {
			t.Parallel()

			assert.True(t, helpers.MatchesAddressPatterns([]string{"0x[^A-F]{40}"}, checksummed))
		})
	})

	t.Run("Given a pattern that is only a prefix of the address", func(t *testing.T) {
		t.Parallel()

		// Patterns are full-matched; a prefix would let one price capture every address sharing it.
		t.Run("Should not match", func(t *testing.T) {
			t.Parallel()

			assert.False(t, helpers.MatchesAddressPatterns([]string{"0xf39f"}, checksummed))
			assert.False(
				t,
				helpers.MatchesAddressPatterns([]string{"f39fd6e51aad88f6f4ce6ab8827279cfffb92266"}, checksummed),
			)
		})
	})

	t.Run("Given the catch-all pattern", func(t *testing.T) {
		t.Parallel()

		t.Run("Should match anyone", func(t *testing.T) {
			t.Parallel()

			assert.True(t, helpers.MatchesAddressPatterns([]string{".*"}, checksummed))
		})
	})

	t.Run("Given an uncompilable pattern", func(t *testing.T) {
		t.Parallel()

		t.Run("Should skip it and keep matching the others", func(t *testing.T) {
			t.Parallel()

			assert.True(t, helpers.MatchesAddressPatterns([]string{"0x[(", strings.ToLower(checksummed)}, checksummed))
		})
	})

	t.Run("Given no patterns at all", func(t *testing.T) {
		t.Parallel()

		t.Run("Should match nobody", func(t *testing.T) {
			t.Parallel()

			assert.False(t, helpers.MatchesAddressPatterns([]string{}, checksummed))
		})
	})
}
