package requests_test

import (
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/requests"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingAppendConversationMessagesRequest(t *testing.T) {
	t.Parallel()

	t.Run("Given the request has at least one message", func(t *testing.T) {
		t.Parallel()
		req := &requests.AppendConversationMessagesRequest{
			Messages: []requests.ConversationMessageInput{
				{
					Role:      "user",
					Content:   "hi",
					Timestamp: time.Now(),
				},
			},
		}

		t.Run("Should pass validation", func(t *testing.T) {
			t.Parallel()

			err := req.ValidateAndSanitize()
			assert.NoError(t, err)
		})
	})

	t.Run("Given the messages slice is empty", func(t *testing.T) {
		t.Parallel()
		req := &requests.AppendConversationMessagesRequest{Messages: nil}

		t.Run("Should reject with a validation error on the messages field", func(t *testing.T) {
			t.Parallel()

			err := req.ValidateAndSanitize()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "messages")
		})
	})
}
