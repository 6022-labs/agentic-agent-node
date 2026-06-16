package requests_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/requests"
	"github.com/stretchr/testify/assert"
)

func TestWhenValidatingSetLlmConfigRequest(t *testing.T) {
	t.Parallel()

	t.Run("Given provider, model, and apiKey are all set", func(t *testing.T) {
		t.Parallel()
		req := &requests.SetLlmConfigRequest{
			Provider: "anthropic",
			Model:    "claude-opus-4-7",
			ApiKey:   "sk-test",
		}

		t.Run("Should pass validation", func(t *testing.T) {
			t.Parallel()
			err := req.ValidateAndSanitize()
			assert.NoError(t, err)
		})
	})

	t.Run("Given provider is whitespace", func(t *testing.T) {
		t.Parallel()
		req := &requests.SetLlmConfigRequest{Provider: "   ", Model: "claude-opus-4-7", ApiKey: "sk-test"}

		t.Run("Should reject with a validation error on provider", func(t *testing.T) {
			t.Parallel()
			err := req.ValidateAndSanitize()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "provider")
		})
	})

	t.Run("Given model is empty", func(t *testing.T) {
		t.Parallel()
		req := &requests.SetLlmConfigRequest{Provider: "anthropic", Model: "", ApiKey: "sk-test"}

		t.Run("Should reject with a validation error on model", func(t *testing.T) {
			t.Parallel()
			err := req.ValidateAndSanitize()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "model")
		})
	})

	t.Run("Given apiKey is empty", func(t *testing.T) {
		t.Parallel()
		req := &requests.SetLlmConfigRequest{Provider: "anthropic", Model: "claude-opus-4-7", ApiKey: ""}

		t.Run("Should pass validation (empty means keep the stored key; runtime enforces first setup)", func(t *testing.T) {
			t.Parallel()
			assert.NoError(t, req.ValidateAndSanitize())
		})
	})
}
