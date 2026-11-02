package models_test

import (
	"strings"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/stretchr/testify/assert"
)

func TestWhenMatchingAPaymentReceiveRulePrice(t *testing.T) {
	t.Parallel()

	// The EIP-55 checksummed form of the address the seeded price stores lowercased.
	checksummed := common_tests.RandomEthAddress().Hex()

	price := func(patterns ...string) *models.PaymentReceiveRulePrice {
		return &models.PaymentReceiveRulePrice{AddressPatterns: patterns}
	}

	t.Run("Given the pattern is the lowercase address the seed stores", func(t *testing.T) {
		t.Parallel()

		t.Run("Should match the caller whatever case they present", func(t *testing.T) {
			t.Parallel()

			assert.True(t, price(strings.ToLower(checksummed)).Matches(checksummed))
			assert.True(t, price(strings.ToLower(checksummed)).Matches(strings.ToLower(checksummed)))
		})
	})

	t.Run("Given the pattern spells the address in checksummed case", func(t *testing.T) {
		t.Parallel()

		// Deliberate: folding case would hand free prices to payers who used to pay.
		t.Run("Should not match", func(t *testing.T) {
			t.Parallel()

			assert.False(t, price(checksummed).Matches(checksummed))
		})
	})

	t.Run("Given a class that negates uppercase hex", func(t *testing.T) {
		t.Parallel()

		// Matching the lowercased address keeps this working; folding case would make
		// Go's negated class exclude a-f as well and reject every address.
		t.Run("Should still match every address", func(t *testing.T) {
			t.Parallel()

			assert.True(t, price("0x[^A-F]{40}").Matches(checksummed))
		})
	})

	t.Run("Given a pattern that is only a prefix of the address", func(t *testing.T) {
		t.Parallel()

		// Patterns are full-matched; a prefix would let one price capture every address sharing it.
		t.Run("Should not match", func(t *testing.T) {
			t.Parallel()

			assert.False(t, price("0xf39f").Matches(checksummed))
			assert.False(t, price("f39fd6e51aad88f6f4ce6ab8827279cfffb92266").Matches(checksummed))
		})
	})

	t.Run("Given the catch-all pattern", func(t *testing.T) {
		t.Parallel()

		t.Run("Should match anyone", func(t *testing.T) {
			t.Parallel()

			assert.True(t, price(".*").Matches(checksummed))
		})
	})

	t.Run("Given an uncompilable pattern", func(t *testing.T) {
		t.Parallel()

		t.Run("Should be skipped rather than match", func(t *testing.T) {
			t.Parallel()

			assert.False(t, price("0x[(").Matches(checksummed))
		})
	})

	t.Run("Given no patterns at all", func(t *testing.T) {
		t.Parallel()

		t.Run("Should match nobody", func(t *testing.T) {
			t.Parallel()

			assert.False(t, price().Matches(checksummed))
		})
	})
}
