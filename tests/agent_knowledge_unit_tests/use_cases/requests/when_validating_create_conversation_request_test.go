package requests_test

import (
	"strings"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/requests"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingCreateConversationRequest(t *testing.T) {
	t.Parallel()

	validCreateConversationRequest := func() *requests.CreateConversationRequest {
		return &requests.CreateConversationRequest{Name: "Audit"}
	}

	t.Run("Given a name padded with whitespace", func(t *testing.T) {
		t.Parallel()

		t.Run("Should collapse it", func(t *testing.T) {
			t.Parallel()

			req := validCreateConversationRequest()
			req.Name = "  ERC20   mint audit "

			assert.NoError(t, req.ValidateAndSanitize())
			assert.Equal(t, "ERC20 mint audit", req.Name)
		})
	})

	t.Run("Given an empty name", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject on the name field", func(t *testing.T) {
			t.Parallel()

			req := validCreateConversationRequest()
			req.Name = "   "

			assert.ErrorContains(t, req.ValidateAndSanitize(), "name")
		})
	})

	t.Run("Given a name longer than the column", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject on the name field", func(t *testing.T) {
			t.Parallel()

			req := validCreateConversationRequest()
			req.Name = strings.Repeat("a", requests.ConversationNameMaxLength+1)

			assert.ErrorContains(t, req.ValidateAndSanitize(), "name")
		})
	})

	t.Run("Given a name of exactly the maximum length", func(t *testing.T) {
		t.Parallel()

		t.Run("Should pass validation", func(t *testing.T) {
			t.Parallel()

			req := validCreateConversationRequest()
			req.Name = strings.Repeat("a", requests.ConversationNameMaxLength)

			assert.NoError(t, req.ValidateAndSanitize())
		})
	})

	t.Run("Given a uuid-shaped name", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject on the uuid rule", func(t *testing.T) {
			t.Parallel()

			req := validCreateConversationRequest()
			req.Name = uuid.NewString()

			assert.ErrorContains(t, req.ValidateAndSanitize(), "uuid")
		})
	})

}
