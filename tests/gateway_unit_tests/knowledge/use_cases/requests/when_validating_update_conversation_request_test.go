package requests_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway/knowledge/use_cases/requests"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingUpdateConversationRequest(t *testing.T) {
	t.Parallel()

	validUpdateConversationRequest := func() *requests.UpdateConversationRequest {
		return &requests.UpdateConversationRequest{
			ConversationIdRequest:   requests.ConversationIdRequest{ConversationId: "conv-1"},
			ConversationNameRequest: requests.ConversationNameRequest{Name: "Trip planning"},
		}
	}

	t.Run("Given a blank conversation id", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject on the conversationId field", func(t *testing.T) {
			t.Parallel()

			req := validUpdateConversationRequest()
			req.ConversationId = "  "

			assert.ErrorContains(t, req.ValidateAndSanitize(), "conversationId")
		})
	})

	t.Run("Given a blank name", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject on the name field", func(t *testing.T) {
			t.Parallel()

			req := validUpdateConversationRequest()
			req.Name = "   "

			assert.ErrorContains(t, req.ValidateAndSanitize(), "name")
		})
	})

	t.Run("Given a fully populated request", func(t *testing.T) {
		t.Parallel()

		t.Run("Should pass validation", func(t *testing.T) {
			t.Parallel()

			assert.NoError(t, validUpdateConversationRequest().ValidateAndSanitize())
		})
	})
}
