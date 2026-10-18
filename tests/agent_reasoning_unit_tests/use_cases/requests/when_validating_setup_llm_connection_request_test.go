package requests_test

import (
	"testing"

	reasoning_enums "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingSetupLLMConnectionRequest(t *testing.T) {
	t.Parallel()

	generateSetupLLMConnectionRequest := func() *requests.SetupLLMConnectionRequest {
		return &requests.SetupLLMConnectionRequest{
			Platform: reasoning_enums.LLMConnectionPlatformOpenAi,
			Model:    "  gpt-4o  ",
			ApiKey:   "  sk-secret  ",
		}
	}

	t.Run("Given the platform is unknown", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError", func(t *testing.T) {
			t.Parallel()

			req := generateSetupLLMConnectionRequest()
			req.Platform = "not-a-platform"

			assert.ErrorContains(t, req.ValidateAndSanitize(), "platform")
		})
	})

	t.Run("Given a blank model", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError", func(t *testing.T) {
			t.Parallel()

			req := generateSetupLLMConnectionRequest()
			req.Model = "   "

			assert.ErrorContains(t, req.ValidateAndSanitize(), "model")
		})
	})

	t.Run("Given a blank api key", func(t *testing.T) {
		t.Parallel()

		t.Run("Should pass validation", func(t *testing.T) {
			t.Parallel()

			req := generateSetupLLMConnectionRequest()
			req.ApiKey = "   "

			assert.NoError(t, req.ValidateAndSanitize())
		})

		t.Run("Should sanitize it to an empty string", func(t *testing.T) {
			t.Parallel()

			req := generateSetupLLMConnectionRequest()
			req.ApiKey = "   "

			_ = req.ValidateAndSanitize()

			assert.Equal(t, "", req.ApiKey)
		})
	})

	t.Run("Given a fully populated request", func(t *testing.T) {
		t.Parallel()

		t.Run("Should sanitize the fields", func(t *testing.T) {
			t.Parallel()

			req := generateSetupLLMConnectionRequest()
			err := req.ValidateAndSanitize()

			assert.NoError(t, err)
			assert.Equal(t, "gpt-4o", req.Model)
			assert.Equal(t, "sk-secret", req.ApiKey)
		})
	})
}
