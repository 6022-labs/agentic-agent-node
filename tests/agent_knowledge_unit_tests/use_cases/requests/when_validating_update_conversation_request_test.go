package requests_test

import (
	"strings"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/requests"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingUpdateConversationRequest(t *testing.T) {
	t.Parallel()

	t.Run("Given the route id is missing", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject on the conversationId field", func(t *testing.T) {
			t.Parallel()

			req := &requests.UpdateConversationRequest{ConversationId: "  ", Name: "Solidity audits"}

			err := req.ValidateAndSanitize()

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "conversationId")
		})
	})

	t.Run("Given an unusable name", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject an empty one", func(t *testing.T) {
			t.Parallel()

			req := &requests.UpdateConversationRequest{ConversationId: "c1", Name: " "}

			err := req.ValidateAndSanitize()

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "name")
		})

		t.Run("Should reject one longer than the column", func(t *testing.T) {
			t.Parallel()

			req := &requests.UpdateConversationRequest{
				ConversationId: "c1",
				Name:           strings.Repeat("a", requests.ConversationNameMaxLength+1),
			}

			assert.Error(t, req.ValidateAndSanitize())
		})

		t.Run("Should reject a uuid-shaped one, so it cannot pass for a session id", func(t *testing.T) {
			t.Parallel()

			req := &requests.UpdateConversationRequest{ConversationId: "c1", Name: uuid.NewString()}

			err := req.ValidateAndSanitize()

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "uuid")
		})
	})

	t.Run("Given a conversation id and a name", func(t *testing.T) {
		t.Parallel()

		t.Run("Should trim both and pass", func(t *testing.T) {
			t.Parallel()

			req := &requests.UpdateConversationRequest{ConversationId: " c1 ", Name: "  Solidity   audits "}

			assert.NoError(t, req.ValidateAndSanitize())
			assert.Equal(t, "c1", req.ConversationId)
			assert.Equal(t, "Solidity audits", req.Name)
		})
	})
}
