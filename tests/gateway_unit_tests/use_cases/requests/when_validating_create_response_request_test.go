package requests_test

import (
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/gateway/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/requests"
	"github.com/stretchr/testify/assert"
)

func validInputMessage() *requests.InputMessageRequest {
	return &requests.InputMessageRequest{
		Id:        "msg-1",
		Author:    "alice.eth",
		Role:      enums.MessageRoleUser,
		Content:   "hello",
		Timestamp: time.Now(),
	}
}

func validCreateResponseRequest() *requests.CreateResponseRequest {
	return &requests.CreateResponseRequest{
		ConversationId: "conv-1",
		Platform:       enums.ConversationPlatformSlack,
		Prompt:         "do the thing",
		Input:          []*requests.InputMessageRequest{validInputMessage()},
	}
}

func TestWhenValidatingCreateResponseRequest(t *testing.T) {
	t.Parallel()

	t.Run("Given a fully populated request", func(t *testing.T) {
		t.Parallel()
		req := validCreateResponseRequest()

		t.Run("Should pass validation", func(t *testing.T) {
			t.Parallel()
			err := req.ValidateAndSanitize()
			assert.NoError(t, err)
		})
	})

	t.Run("Given conversationId is whitespace", func(t *testing.T) {
		t.Parallel()
		req := validCreateResponseRequest()
		req.ConversationId = "   "

		t.Run("Should reject with a validation error on conversationId", func(t *testing.T) {
			t.Parallel()
			err := req.ValidateAndSanitize()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "conversationId")
		})
	})

	t.Run("Given prompt is empty", func(t *testing.T) {
		t.Parallel()
		req := validCreateResponseRequest()
		req.Prompt = ""

		t.Run("Should reject with a validation error on prompt", func(t *testing.T) {
			t.Parallel()
			err := req.ValidateAndSanitize()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "prompt")
		})
	})

	t.Run("Given platform is unknown", func(t *testing.T) {
		t.Parallel()
		req := validCreateResponseRequest()
		req.Platform = "telepathy"

		t.Run("Should fall back to ConversationPlatformNone without erroring", func(t *testing.T) {
			t.Parallel()
			err := req.ValidateAndSanitize()
			assert.NoError(t, err)
			assert.Equal(t, enums.ConversationPlatformNone, req.Platform)
		})
	})

	t.Run("Given an input message has an empty id", func(t *testing.T) {
		t.Parallel()
		req := validCreateResponseRequest()
		req.Input[0].Id = ""

		t.Run("Should reject with a validation error referencing the index", func(t *testing.T) {
			t.Parallel()
			err := req.ValidateAndSanitize()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "input.id")
			assert.Contains(t, err.Error(), "0")
		})
	})

	t.Run("Given an input message has empty content", func(t *testing.T) {
		t.Parallel()
		req := validCreateResponseRequest()
		req.Input[0].Content = "  "

		t.Run("Should reject", func(t *testing.T) {
			t.Parallel()
			err := req.ValidateAndSanitize()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "input.content")
		})
	})

	t.Run("Given an input message has a zero timestamp", func(t *testing.T) {
		t.Parallel()
		req := validCreateResponseRequest()
		req.Input[0].Timestamp = time.Time{}

		t.Run("Should reject", func(t *testing.T) {
			t.Parallel()
			err := req.ValidateAndSanitize()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "input.timestamp")
		})
	})

	t.Run("Given an input message has an unknown role", func(t *testing.T) {
		t.Parallel()
		req := validCreateResponseRequest()
		req.Input[0].Role = "wizard"

		t.Run("Should reject and list the allowed roles", func(t *testing.T) {
			t.Parallel()
			err := req.ValidateAndSanitize()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "input.role")
		})
	})

	t.Run("Given a tool has an empty name", func(t *testing.T) {
		t.Parallel()
		req := validCreateResponseRequest()
		req.Tools = []*requests.InputToolRequest{{Name: "  "}}

		t.Run("Should reject with a validation error on tools.name", func(t *testing.T) {
			t.Parallel()
			err := req.ValidateAndSanitize()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "tools.name")
		})
	})

	t.Run("Given a tool sets endpoint without url", func(t *testing.T) {
		t.Parallel()
		req := validCreateResponseRequest()
		req.Tools = []*requests.InputToolRequest{{
			Name:     "search",
			Endpoint: &requests.InputToolEndpointRequest{Url: ""},
		}}

		t.Run("Should reject with a validation error on tools.endpoint.url", func(t *testing.T) {
			t.Parallel()
			err := req.ValidateAndSanitize()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "tools.endpoint.url")
		})
	})
}
