package requests_test

import (
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingCreateResponseRequest(t *testing.T) {
	t.Parallel()

	now := time.Now()

	t.Run("Given a valid request", func(t *testing.T) {
		t.Parallel()

		req := &requests.CreateResponseRequest{
			ConversationId: "conv-1",
			Inputs: []requests.CreateResponseRequestInput{
				{Role: "user", Content: "hi", Timestamp: now},
			},
		}

		t.Run("Should pass validation and trim conversationId", func(t *testing.T) {
			t.Parallel()

			err := req.ValidateAndSanitize()
			assert.NoError(t, err)
			assert.Equal(t, "conv-1", req.ConversationId)
		})
	})

	t.Run("Given a conversationId that is whitespace", func(t *testing.T) {
		t.Parallel()

		req := &requests.CreateResponseRequest{
			ConversationId: "   ",
			Inputs: []requests.CreateResponseRequestInput{
				{Role: "user", Content: "hi", Timestamp: now},
			},
		}

		t.Run("Should reject with a conversationId validation error", func(t *testing.T) {
			t.Parallel()

			err := req.ValidateAndSanitize()
			assert.ErrorContains(t, err, "conversationId")
		})
	})

	t.Run("Given an empty inputs slice", func(t *testing.T) {
		t.Parallel()

		req := &requests.CreateResponseRequest{ConversationId: "conv-1"}

		t.Run("Should reject", func(t *testing.T) {
			t.Parallel()

			err := req.ValidateAndSanitize()
			assert.ErrorContains(t, err, "inputs")
		})
	})

	t.Run("Given an input message with empty role", func(t *testing.T) {
		t.Parallel()

		req := &requests.CreateResponseRequest{
			ConversationId: "conv-1",
			Inputs: []requests.CreateResponseRequestInput{
				{Role: "", Content: "hi", Timestamp: now},
			},
		}

		t.Run("Should reject", func(t *testing.T) {
			t.Parallel()

			err := req.ValidateAndSanitize()
			assert.ErrorContains(t, err, "inputs.role")
		})
	})

	t.Run("Given an input message with empty content", func(t *testing.T) {
		t.Parallel()

		req := &requests.CreateResponseRequest{
			ConversationId: "conv-1",
			Inputs: []requests.CreateResponseRequestInput{
				{Role: "user", Content: "  ", Timestamp: now},
			},
		}

		t.Run("Should reject", func(t *testing.T) {
			t.Parallel()

			err := req.ValidateAndSanitize()
			assert.ErrorContains(t, err, "inputs.content")
		})
	})

	t.Run("Given an input message with zero timestamp", func(t *testing.T) {
		t.Parallel()

		req := &requests.CreateResponseRequest{
			ConversationId: "conv-1",
			Inputs: []requests.CreateResponseRequestInput{
				{Role: "user", Content: "hi", Timestamp: time.Time{}},
			},
		}

		t.Run("Should reject", func(t *testing.T) {
			t.Parallel()

			err := req.ValidateAndSanitize()
			assert.ErrorContains(t, err, "inputs.timestamp")
		})
	})
}
