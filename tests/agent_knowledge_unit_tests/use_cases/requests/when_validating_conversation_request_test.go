package requests_test

import (
	"strings"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/requests"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingConversationRequest(t *testing.T) {
	t.Parallel()

	t.Run("Given a padded conversation id", func(t *testing.T) {
		t.Parallel()

		t.Run("Should trim it", func(t *testing.T) {
			t.Parallel()

			req := &requests.ConversationRequest{ConversationId: " conv-1 "}

			assert.NoError(t, req.ValidateAndSanitize())
			assert.Equal(t, "conv-1", req.ConversationId)
		})
	})

	t.Run("Given a blank conversation id", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject on the conversationId field", func(t *testing.T) {
			t.Parallel()

			req := &requests.ConversationRequest{ConversationId: "  "}

			assert.ErrorContains(t, req.ValidateAndSanitize(), "conversationId")
		})
	})

	t.Run("Given a conversation id longer than the column", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject on the conversationId field", func(t *testing.T) {
			t.Parallel()

			req := &requests.ConversationRequest{
				ConversationId: strings.Repeat("k", requests.ConversationIdMaxLength+1),
			}

			assert.ErrorContains(t, req.ValidateAndSanitize(), "conversationId")
		})
	})

	t.Run("Given a conversation id of exactly the maximum length", func(t *testing.T) {
		t.Parallel()

		t.Run("Should pass validation", func(t *testing.T) {
			t.Parallel()

			req := &requests.ConversationRequest{ConversationId: strings.Repeat("é", requests.ConversationIdMaxLength)}

			assert.NoError(t, req.ValidateAndSanitize())
		})
	})
}
