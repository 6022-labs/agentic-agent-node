package models_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_inference/models"
	"github.com/stretchr/testify/assert"
)

func TestWhenParsingCompletionsToolCallArguments(t *testing.T) {
	t.Parallel()

	t.Run("Given a malformed JSON object", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a nil map", func(t *testing.T) {
			t.Parallel()

			function := &models.ExternalOpenAICompletionsStandardToolCallFunction{Arguments: "{bad"}

			args, _ := function.ParsedArguments()

			assert.Nil(t, args)
		})

		t.Run("Should return an error", func(t *testing.T) {
			t.Parallel()

			function := &models.ExternalOpenAICompletionsStandardToolCallFunction{Arguments: "{bad"}

			_, err := function.ParsedArguments()

			assert.ErrorContains(t, err, "invalid character")
		})
	})

	t.Run("Given a blank arguments string", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a non-nil empty map without error", func(t *testing.T) {
			t.Parallel()

			for _, value := range []string{"", "   "} {
				function := &models.ExternalOpenAICompletionsStandardToolCallFunction{Arguments: value}

				args, err := function.ParsedArguments()

				assert.NoError(t, err)
				assert.NotNil(t, args)
				assert.Empty(t, args)
			}
		})
	})

	t.Run("Given an empty JSON object", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return an empty map without error", func(t *testing.T) {
			t.Parallel()

			function := &models.ExternalOpenAICompletionsStandardToolCallFunction{Arguments: "{}"}

			args, err := function.ParsedArguments()

			assert.NoError(t, err)
			assert.NotNil(t, args)
			assert.Empty(t, args)
		})
	})

	t.Run("Given a populated JSON object", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return the parsed arguments without error", func(t *testing.T) {
			t.Parallel()

			function := &models.ExternalOpenAICompletionsStandardToolCallFunction{Arguments: `{"city":"paris"}`}

			args, err := function.ParsedArguments()

			assert.NoError(t, err)
			assert.Equal(t, map[string]any{"city": "paris"}, args)
		})
	})
}
