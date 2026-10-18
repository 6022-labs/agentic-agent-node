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

	validUpdateConversationRequest := func() *requests.UpdateConversationRequest {
		return &requests.UpdateConversationRequest{
			ConversationRequest: requests.ConversationRequest{ConversationId: "c1"},
			Name:                "Solidity audits",
		}
	}

	t.Run("Given a fully populated request", func(t *testing.T) {
		t.Parallel()

		t.Run("Should pass validation", func(t *testing.T) {
			t.Parallel()

			assert.NoError(t, validUpdateConversationRequest().ValidateAndSanitize())
		})
	})

	t.Run("Given a blank conversation id", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject on the conversationId field", func(t *testing.T) {
			t.Parallel()

			req := validUpdateConversationRequest()
			req.ConversationId = "  "

			assert.ErrorContains(t, req.ValidateAndSanitize(), "conversationId")
		})
	})

	t.Run("Given a name padded with whitespace", func(t *testing.T) {
		t.Parallel()

		t.Run("Should collapse it", func(t *testing.T) {
			t.Parallel()

			req := validUpdateConversationRequest()
			req.Name = "  Solidity   audits "

			assert.NoError(t, req.ValidateAndSanitize())
			assert.Equal(t, "Solidity audits", req.Name)
		})
	})

	t.Run("Given an empty name", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject on the name field", func(t *testing.T) {
			t.Parallel()

			req := validUpdateConversationRequest()
			req.Name = " "

			assert.ErrorContains(t, req.ValidateAndSanitize(), "name")
		})
	})

	t.Run("Given a name longer than the column", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject on the name field", func(t *testing.T) {
			t.Parallel()

			req := validUpdateConversationRequest()
			req.Name = strings.Repeat("a", requests.ConversationNameMaxLength+1)

			assert.ErrorContains(t, req.ValidateAndSanitize(), "name")
		})
	})

	t.Run("Given a name of exactly the maximum length", func(t *testing.T) {
		t.Parallel()

		t.Run("Should pass validation", func(t *testing.T) {
			t.Parallel()

			req := validUpdateConversationRequest()
			req.Name = strings.Repeat("a", requests.ConversationNameMaxLength)

			assert.NoError(t, req.ValidateAndSanitize())
		})
	})

	t.Run("Given a uuid-shaped name", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject on the uuid rule", func(t *testing.T) {
			t.Parallel()

			req := validUpdateConversationRequest()
			req.Name = uuid.NewString()

			assert.ErrorContains(t, req.ValidateAndSanitize(), "uuid")
		})
	})

}
