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
			Platform: reasoning_enums.LLMConnectionPlatformOpenAI,
			Model:    "  gpt-4o  ",
			ApiKey:   "  sk-secret  ",
		}
	}

	t.Run("Given the request is valid", func(t *testing.T) {
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

	t.Run("Given the platform is unknown", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error", func(t *testing.T) {
			t.Parallel()

			req := generateSetupLLMConnectionRequest()
			req.Platform = "not-a-platform"

			assert.Error(t, req.ValidateAndSanitize())
		})
	})

	t.Run("Given the model is empty", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error", func(t *testing.T) {
			t.Parallel()

			req := generateSetupLLMConnectionRequest()
			req.Model = "   "

			assert.Error(t, req.ValidateAndSanitize())
		})
	})

	t.Run("Given the api key is empty", func(t *testing.T) {
		t.Parallel()

		t.Run("Should pass validation and sanitize to an empty string", func(t *testing.T) {
			t.Parallel()

			req := generateSetupLLMConnectionRequest()
			req.ApiKey = "   "

			assert.NoError(t, req.ValidateAndSanitize())
			assert.Equal(t, "", req.ApiKey)
		})
	})
}
