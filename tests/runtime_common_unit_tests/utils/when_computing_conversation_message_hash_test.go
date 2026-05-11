package utils_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/runtime_common/utils"
	"github.com/stretchr/testify/assert"
)

func TestWhenComputingConversationMessageHash(t *testing.T) {
	t.Parallel()

	t.Run("Given identical inputs", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return the same hash", func(t *testing.T) {
			t.Parallel()

			h1, err1 := utils.ComputeConversationMessageHash("user", "alice", "slack", "hello", "conv-1")
			h2, err2 := utils.ComputeConversationMessageHash("user", "alice", "slack", "hello", "conv-1")

			assert.NoError(t, err1)
			assert.NoError(t, err2)
			assert.Equal(t, h1, h2)
			assert.Len(t, h1, 128) // sha-512 hex
		})
	})

	t.Run("Given different content", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return different hashes", func(t *testing.T) {
			t.Parallel()

			h1, _ := utils.ComputeConversationMessageHash("user", "alice", "slack", "hello", "conv-1")
			h2, _ := utils.ComputeConversationMessageHash("user", "alice", "slack", "world", "conv-1")

			assert.NotEqual(t, h1, h2)
		})
	})

	t.Run("Given different conversationId", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return different hashes", func(t *testing.T) {
			t.Parallel()

			h1, _ := utils.ComputeConversationMessageHash("user", "alice", "slack", "hello", "conv-1")
			h2, _ := utils.ComputeConversationMessageHash("user", "alice", "slack", "hello", "conv-2")

			assert.NotEqual(t, h1, h2)
		})
	})

	t.Run("Given different role", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return different hashes", func(t *testing.T) {
			t.Parallel()

			h1, _ := utils.ComputeConversationMessageHash("user", "alice", "slack", "hello", "conv-1")
			h2, _ := utils.ComputeConversationMessageHash("assistant", "alice", "slack", "hello", "conv-1")

			assert.NotEqual(t, h1, h2)
		})
	})

	t.Run("Given different author", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return different hashes", func(t *testing.T) {
			t.Parallel()

			h1, _ := utils.ComputeConversationMessageHash("user", "alice", "slack", "hello", "conv-1")
			h2, _ := utils.ComputeConversationMessageHash("user", "bob", "slack", "hello", "conv-1")

			assert.NotEqual(t, h1, h2)
		})
	})

	t.Run("Given different platform", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return different hashes", func(t *testing.T) {
			t.Parallel()

			h1, _ := utils.ComputeConversationMessageHash("user", "alice", "slack", "hello", "conv-1")
			h2, _ := utils.ComputeConversationMessageHash("user", "alice", "discord", "hello", "conv-1")

			assert.NotEqual(t, h1, h2)
		})
	})

	t.Run("Given the empty platform", func(t *testing.T) {
		t.Parallel()

		t.Run("Should still produce a deterministic hash distinct from any non-empty platform", func(t *testing.T) {
			t.Parallel()

			h1, _ := utils.ComputeConversationMessageHash("user", "alice", "", "hello", "conv-1")
			h2, _ := utils.ComputeConversationMessageHash("user", "alice", "", "hello", "conv-1")
			withPlatform, _ := utils.ComputeConversationMessageHash("user", "alice", "slack", "hello", "conv-1")

			assert.Equal(t, h1, h2)
			assert.NotEqual(t, h1, withPlatform)
		})
	})

	t.Run("Given an emoji glyph vs its demojized shortcode form", func(t *testing.T) {
		t.Parallel()

		t.Run("Should differ — caller is responsible for demojizing first", func(t *testing.T) {
			t.Parallel()

			withGlyph, _ := utils.ComputeConversationMessageHash("user", "alice", "slack", "hi 🔥", "conv-1")
			withShortcode, _ := utils.ComputeConversationMessageHash("user", "alice", "slack", "hi :fire:", "conv-1")

			assert.NotEqual(t, withGlyph, withShortcode,
				"hash compare requires content already demojized — this confirms reasoning + knowledge MUST demojize before hashing")
		})
	})
}
