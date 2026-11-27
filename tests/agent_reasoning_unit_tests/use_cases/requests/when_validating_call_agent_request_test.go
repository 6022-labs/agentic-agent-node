package requests_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingCallAgentRequest(t *testing.T) {
	t.Parallel()

	generateCallAgentRequest := func() *requests.CallAgentRequest {
		return &requests.CallAgentRequest{
			Url:              " https://peer.example/ ",
			Message:          " hello ",
			ContextId:        " conversation-1 ",
			PaymentSignature: " eyJ4NDAy ",
		}
	}

	t.Run("Given a blank url", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject on the url field", func(t *testing.T) {
			t.Parallel()

			req := generateCallAgentRequest()
			req.Url = "   "

			var validationError *apperrors.ValidationError
			assert.ErrorAs(t, req.ValidateAndSanitize(), &validationError)
			assert.Equal(t, "url", validationError.Field)
		})
	})

	t.Run("Given a url that is not a url", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject on the url field", func(t *testing.T) {
			t.Parallel()

			req := generateCallAgentRequest()
			req.Url = "https://peer example/%zz"

			var validationError *apperrors.ValidationError
			assert.ErrorAs(t, req.ValidateAndSanitize(), &validationError)
			assert.Equal(t, "url", validationError.Field)
		})
	})

	t.Run("Given a url with another scheme", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject on the url field", func(t *testing.T) {
			t.Parallel()

			req := generateCallAgentRequest()
			req.Url = "file:///etc/passwd"

			var validationError *apperrors.ValidationError
			assert.ErrorAs(t, req.ValidateAndSanitize(), &validationError)
			assert.Equal(t, "url", validationError.Field)
		})
	})

	t.Run("Given a url without a host", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject on the url field", func(t *testing.T) {
			t.Parallel()

			req := generateCallAgentRequest()
			req.Url = "https://"

			var validationError *apperrors.ValidationError
			assert.ErrorAs(t, req.ValidateAndSanitize(), &validationError)
			assert.Equal(t, "url", validationError.Field)
		})
	})

	t.Run("Given a url carrying credentials", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject on the url field", func(t *testing.T) {
			t.Parallel()

			req := generateCallAgentRequest()
			req.Url = "https://user:secret@peer.example"

			var validationError *apperrors.ValidationError
			assert.ErrorAs(t, req.ValidateAndSanitize(), &validationError)
			assert.Equal(t, "url", validationError.Field)
		})
	})

	t.Run("Given a url carrying a query", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject on the url field", func(t *testing.T) {
			t.Parallel()

			req := generateCallAgentRequest()
			req.Url = "https://peer.example/internal?"

			var validationError *apperrors.ValidationError
			assert.ErrorAs(t, req.ValidateAndSanitize(), &validationError)
			assert.Equal(t, "url", validationError.Field)
		})
	})

	t.Run("Given a url carrying a fragment", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject on the url field", func(t *testing.T) {
			t.Parallel()

			req := generateCallAgentRequest()
			req.Url = "https://peer.example#"

			var validationError *apperrors.ValidationError
			assert.ErrorAs(t, req.ValidateAndSanitize(), &validationError)
			assert.Equal(t, "url", validationError.Field)
		})
	})

	t.Run("Given a blank message", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject on the message field", func(t *testing.T) {
			t.Parallel()

			req := generateCallAgentRequest()
			req.Message = "   "

			var validationError *apperrors.ValidationError
			assert.ErrorAs(t, req.ValidateAndSanitize(), &validationError)
			assert.Equal(t, "message", validationError.Field)
		})
	})

	t.Run("Given a valid request", func(t *testing.T) {
		t.Parallel()

		t.Run("Should trim every field and the url's trailing slash", func(t *testing.T) {
			t.Parallel()

			req := generateCallAgentRequest()

			assert.NoError(t, req.ValidateAndSanitize())
			assert.Equal(t, &requests.CallAgentRequest{
				Url:              "https://peer.example",
				Message:          "hello",
				ContextId:        "conversation-1",
				PaymentSignature: "eyJ4NDAy",
			}, req)
		})
	})
}
