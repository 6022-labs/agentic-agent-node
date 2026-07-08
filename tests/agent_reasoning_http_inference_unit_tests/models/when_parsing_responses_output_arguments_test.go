package models_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_inference/models"
	"github.com/stretchr/testify/assert"
)

func TestWhenParsingResponsesOutputArguments(t *testing.T) {
	t.Parallel()

	stringPtr := func(value string) *string { return &value }

	t.Run("Given nil arguments", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a non-nil empty map without error", func(t *testing.T) {
			t.Parallel()

			output := &models.ExternalOpenAIResponsesStandardResponseOutput{Arguments: nil}

			args, err := output.ParsedArguments()

			assert.NoError(t, err)
			assert.NotNil(t, args)
			assert.Empty(t, args)
		})
	})

	t.Run("Given an empty arguments string", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a non-nil empty map without error", func(t *testing.T) {
			t.Parallel()

			output := &models.ExternalOpenAIResponsesStandardResponseOutput{Arguments: stringPtr("")}

			args, err := output.ParsedArguments()

			assert.NoError(t, err)
			assert.NotNil(t, args)
			assert.Empty(t, args)
		})
	})

	t.Run("Given a whitespace-only arguments string", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a non-nil empty map without error", func(t *testing.T) {
			t.Parallel()

			output := &models.ExternalOpenAIResponsesStandardResponseOutput{Arguments: stringPtr("   ")}

			args, err := output.ParsedArguments()

			assert.NoError(t, err)
			assert.NotNil(t, args)
			assert.Empty(t, args)
		})
	})

	t.Run("Given an empty JSON object", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return an empty map without error", func(t *testing.T) {
			t.Parallel()

			output := &models.ExternalOpenAIResponsesStandardResponseOutput{Arguments: stringPtr("{}")}

			args, err := output.ParsedArguments()

			assert.NoError(t, err)
			assert.NotNil(t, args)
			assert.Empty(t, args)
		})
	})

	t.Run("Given a populated JSON object", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return the parsed arguments without error", func(t *testing.T) {
			t.Parallel()

			output := &models.ExternalOpenAIResponsesStandardResponseOutput{Arguments: stringPtr(`{"city":"paris"}`)}

			args, err := output.ParsedArguments()

			assert.NoError(t, err)
			assert.Equal(t, map[string]any{"city": "paris"}, args)
		})
	})

	t.Run("Given a malformed JSON object", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a nil map and an error", func(t *testing.T) {
			t.Parallel()

			output := &models.ExternalOpenAIResponsesStandardResponseOutput{Arguments: stringPtr("{bad")}

			args, err := output.ParsedArguments()

			assert.Error(t, err)
			assert.Nil(t, args)
		})
	})
}
