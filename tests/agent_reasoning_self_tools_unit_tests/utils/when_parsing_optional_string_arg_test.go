package utils_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
	"github.com/stretchr/testify/assert"
)

func TestWhenParsingOptionalStringArg(t *testing.T) {
	t.Parallel()

	t.Run("Given the argument is missing", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return an empty string", func(t *testing.T) {
			t.Parallel()

			value := utils.OptionalStringArg(nil, "method")

			assert.Empty(t, value)
		})
	})

	t.Run("Given the argument is present", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return its value", func(t *testing.T) {
			t.Parallel()

			value := utils.OptionalStringArg(map[string]any{"method": "x402"}, "method")

			assert.Equal(t, "x402", value)
		})
	})
}
