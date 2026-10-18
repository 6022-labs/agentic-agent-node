package requests_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/requests"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingA2ARequest(t *testing.T) {
	t.Parallel()

	validA2ARequest := func() *requests.A2ARequest {
		return &requests.A2ARequest{
			JsonRpc: "2.0",
			Method:  requests.A2AMethodSendMessage,
			Params: requests.A2AMessageSendParams{
				Message: requests.A2AMessage{Role: "ROLE_USER", Parts: []requests.A2APart{{Text: "hi"}}},
			},
		}
	}

	t.Run("Given an unsupported method", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject on the method field", func(t *testing.T) {
			t.Parallel()

			req := validA2ARequest()
			req.Method = "GetTask"

			assert.ErrorContains(t, req.ValidateAndSanitize(), "method")
		})
	})

	t.Run("Given a message without text", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject on the params.message field", func(t *testing.T) {
			t.Parallel()

			req := validA2ARequest()
			req.Params.Message.Parts = []requests.A2APart{{Text: "  "}}

			assert.ErrorContains(t, req.ValidateAndSanitize(), "params.message")
		})
	})

	t.Run("Given a padded contextId", func(t *testing.T) {
		t.Parallel()

		t.Run("Should trim it", func(t *testing.T) {
			t.Parallel()

			req := validA2ARequest()
			req.Params.Message.ContextId = " ctx-1 "

			assert.NoError(t, req.ValidateAndSanitize())
			assert.Equal(t, "ctx-1", req.Params.Message.ContextId)
		})
	})

	t.Run("Given a fully populated request", func(t *testing.T) {
		t.Parallel()

		t.Run("Should pass validation", func(t *testing.T) {
			t.Parallel()

			assert.NoError(t, validA2ARequest().ValidateAndSanitize())
		})
	})
}
