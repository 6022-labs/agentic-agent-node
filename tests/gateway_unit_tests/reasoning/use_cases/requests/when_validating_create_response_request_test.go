package requests_test

import (
	"encoding/json"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/requests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	t.Run("Given a fully populated request", func(t *testing.T) {
		t.Parallel()
		req := validCreateResponseRequest()

		t.Run("Should pass validation", func(t *testing.T) {
			t.Parallel()
			err := req.ValidateAndSanitize()
			assert.NoError(t, err)
		})
	})

	t.Run("Given input has no messages", func(t *testing.T) {
		t.Parallel()
		req := validCreateResponseRequest()
		req.Input.Messages = nil

		t.Run("Should reject with a validation error on input", func(t *testing.T) {
			t.Parallel()
			err := req.ValidateAndSanitize()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "input")
		})
	})

	t.Run("Given an input message has empty content", func(t *testing.T) {
		t.Parallel()
		req := validCreateResponseRequest()
		req.Input.Messages[0].Content = "  "

		t.Run("Should reject referencing the index", func(t *testing.T) {
			t.Parallel()
			err := req.ValidateAndSanitize()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "input")
			assert.Contains(t, err.Error(), "0")
		})
	})
}

func TestWhenUnmarshallingCreateResponseRequest(t *testing.T) {
	t.Parallel()

	t.Run("Given input is a bare string", func(t *testing.T) {
		t.Parallel()
		var req requests.CreateResponseRequest
		require.NoError(t, json.Unmarshal([]byte(`{"input":"hello"}`), &req))

		t.Run("Should map to a single user message", func(t *testing.T) {
			t.Parallel()
			require.Len(t, req.Input.Messages, 1)
			assert.Equal(t, "user", req.Input.Messages[0].Role)
			assert.Equal(t, "hello", req.Input.Messages[0].Content)
		})
	})

	t.Run("Given input is an array of message items with string content", func(t *testing.T) {
		t.Parallel()
		var req requests.CreateResponseRequest
		require.NoError(t, json.Unmarshal([]byte(`{"input":[{"type":"message","role":"user","content":"hi"}]}`), &req))

		t.Run("Should map role and content", func(t *testing.T) {
			t.Parallel()
			require.Len(t, req.Input.Messages, 1)
			assert.Equal(t, "user", req.Input.Messages[0].Role)
			assert.Equal(t, "hi", req.Input.Messages[0].Content)
		})
	})

	t.Run("Given input content is an array of text parts", func(t *testing.T) {
		t.Parallel()
		var req requests.CreateResponseRequest
		require.NoError(t, json.Unmarshal([]byte(`{"input":[{"role":"user","content":[{"type":"input_text","text":"a"},{"type":"input_text","text":"b"}]}]}`), &req))

		t.Run("Should concatenate the parts", func(t *testing.T) {
			t.Parallel()
			require.Len(t, req.Input.Messages, 1)
			assert.Equal(t, "ab", req.Input.Messages[0].Content)
		})
	})

	t.Run("Given conversation is a bare string", func(t *testing.T) {
		t.Parallel()
		var req requests.CreateResponseRequest
		require.NoError(t, json.Unmarshal([]byte(`{"input":"hi","conversation":"conv-9"}`), &req))

		t.Run("Should resolve the conversation key", func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, "conv-9", req.ConversationKey())
		})
	})

	t.Run("Given conversation is an object", func(t *testing.T) {
		t.Parallel()
		var req requests.CreateResponseRequest
		require.NoError(t, json.Unmarshal([]byte(`{"input":"hi","conversation":{"id":"conv-7"}}`), &req))

		t.Run("Should resolve the conversation key", func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, "conv-7", req.ConversationKey())
		})
	})

	t.Run("Given both previous_response_id and conversation are set", func(t *testing.T) {
		t.Parallel()
		var req requests.CreateResponseRequest
		require.NoError(t, json.Unmarshal([]byte(`{"input":"hi","previous_response_id":"resp-1","conversation":"conv-7"}`), &req))

		t.Run("Should prefer previous_response_id", func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, "resp-1", req.ConversationKey())
		})
	})
}
