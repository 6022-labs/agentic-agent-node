package requests_test

import (
	"strings"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/requests"
	"github.com/stretchr/testify/assert"
)

func validCreateResponseRequest() *requests.CreateResponseRequest {
	return &requests.CreateResponseRequest{
		PreviousResponseId: "conv-1",
		Input: requests.ResponseInput{
			Messages: []requests.InputMessage{{Role: "user", Content: "hello"}},
		},
	}
}

func TestWhenValidatingCreateResponseRequest(t *testing.T) {
	t.Parallel()

	t.Run("Given input has no messages", func(t *testing.T) {
		t.Parallel()
		req := validCreateResponseRequest()
		req.Input.Messages = nil

		t.Run("Should reject on the input field", func(t *testing.T) {
			t.Parallel()
			assert.ErrorContains(t, req.ValidateAndSanitize(), "input")
		})
	})

	t.Run("Given an input message with blank content", func(t *testing.T) {
		t.Parallel()
		req := validCreateResponseRequest()
		req.Input.Messages[0].Content = "  "

		t.Run("Should reject on the input field", func(t *testing.T) {
			t.Parallel()
			err := req.ValidateAndSanitize()
			assert.ErrorContains(t, err, "input")
			assert.ErrorContains(t, err, "index 0")
		})
	})

	t.Run("Given a previous_response_id longer than the column", func(t *testing.T) {
		t.Parallel()
		req := validCreateResponseRequest()
		req.PreviousResponseId = strings.Repeat("k", requests.ConversationKeyMaxLength+1)

		t.Run("Should reject on the conversation field", func(t *testing.T) {
			t.Parallel()
			assert.ErrorContains(t, req.ValidateAndSanitize(), "conversation")
		})
	})

	t.Run("Given a conversation.id longer than the column", func(t *testing.T) {
		t.Parallel()
		req := validCreateResponseRequest()
		req.PreviousResponseId = ""
		req.Conversation = &requests.ConversationRef{Id: strings.Repeat("k", requests.ConversationKeyMaxLength+1)}

		t.Run("Should reject on the conversation field", func(t *testing.T) {
			t.Parallel()
			assert.ErrorContains(t, req.ValidateAndSanitize(), "conversation")
		})
	})

	t.Run("Given a conversation key of exactly the maximum length", func(t *testing.T) {
		t.Parallel()
		req := validCreateResponseRequest()
		req.PreviousResponseId = strings.Repeat("k", requests.ConversationKeyMaxLength)

		t.Run("Should pass validation", func(t *testing.T) {
			t.Parallel()
			assert.NoError(t, req.ValidateAndSanitize())
		})
	})

	t.Run("Given a multibyte conversation key of exactly the maximum length", func(t *testing.T) {
		t.Parallel()
		req := validCreateResponseRequest()
		req.PreviousResponseId = strings.Repeat("é", requests.ConversationKeyMaxLength)

		t.Run("Should pass validation", func(t *testing.T) {
			t.Parallel()
			assert.NoError(t, req.ValidateAndSanitize())
		})
	})

	t.Run("Given a fully populated request", func(t *testing.T) {
		t.Parallel()
		req := validCreateResponseRequest()

		t.Run("Should pass validation", func(t *testing.T) {
			t.Parallel()
			assert.NoError(t, req.ValidateAndSanitize())
		})
	})

}
