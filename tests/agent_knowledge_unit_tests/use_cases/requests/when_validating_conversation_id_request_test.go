package requests_test

import (
	"strings"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingConversationIdRequest(t *testing.T) {
	t.Parallel()

	t.Run("Given a blank conversation id", func(t *testing.T) {
		t.Parallel()
		req := &requests.ConversationIdRequest{ConversationId: "  "}

		t.Run("Should reject on the conversationId field", func(t *testing.T) {
			t.Parallel()

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, req.ValidateAndSanitize(), &validationErr)
			assert.Equal(t, "conversationId", validationErr.Field)
		})
	})

	t.Run("Given a conversation id longer than the column", func(t *testing.T) {
		t.Parallel()
		req := &requests.ConversationIdRequest{
			ConversationId: strings.Repeat("k", requests.ConversationIdMaxLength+1),
		}

		t.Run("Should reject on the conversationId field", func(t *testing.T) {
			t.Parallel()

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, req.ValidateAndSanitize(), &validationErr)
			assert.Equal(t, "conversationId", validationErr.Field)
		})
	})

	t.Run("Given a padded conversation id", func(t *testing.T) {
		t.Parallel()
		req := &requests.ConversationIdRequest{ConversationId: " conv-1 "}

		t.Run("Should trim it", func(t *testing.T) {
			t.Parallel()

			assert.NoError(t, req.ValidateAndSanitize())
			assert.Equal(t, "conv-1", req.ConversationId)
		})
	})

	t.Run("Given a conversation id of exactly the maximum length", func(t *testing.T) {
		t.Parallel()
		req := &requests.ConversationIdRequest{
			ConversationId: strings.Repeat("é", requests.ConversationIdMaxLength),
		}

		t.Run("Should pass validation", func(t *testing.T) {
			t.Parallel()

			assert.NoError(t, req.ValidateAndSanitize())
		})
	})
}
