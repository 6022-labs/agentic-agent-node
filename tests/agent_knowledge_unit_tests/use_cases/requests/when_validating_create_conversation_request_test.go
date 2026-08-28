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

	t.Run("Given a name padded with whitespace", func(t *testing.T) {
		t.Parallel()

		t.Run("Should collapse it and pass", func(t *testing.T) {
			t.Parallel()

			req := &requests.CreateConversationRequest{Name: "  ERC20   mint audit "}

			assert.NoError(t, req.ValidateAndSanitize())
			assert.Equal(t, "ERC20 mint audit", req.Name)
		})
	})

	t.Run("Given an empty name", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject on the name field", func(t *testing.T) {
			t.Parallel()

			req := &requests.CreateConversationRequest{Name: "   "}

			err := req.ValidateAndSanitize()

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "name")
		})
	})

	t.Run("Given a name longer than the column", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject it", func(t *testing.T) {
			t.Parallel()

			req := &requests.CreateConversationRequest{
				Name: strings.Repeat("a", requests.ConversationNameMaxLength+1),
			}

			assert.Error(t, req.ValidateAndSanitize())
		})

		t.Run("Should accept a name of exactly the maximum length", func(t *testing.T) {
			t.Parallel()

			req := &requests.CreateConversationRequest{
				Name: strings.Repeat("a", requests.ConversationNameMaxLength),
			}

			assert.NoError(t, req.ValidateAndSanitize())
		})
	})

	t.Run("Given a uuid-shaped name", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject it, so it cannot pass for a session id", func(t *testing.T) {
			t.Parallel()

			req := &requests.CreateConversationRequest{Name: uuid.NewString()}

			err := req.ValidateAndSanitize()

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "uuid")
		})
	})
}
