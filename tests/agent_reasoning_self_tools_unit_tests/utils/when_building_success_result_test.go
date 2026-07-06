package utils_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
	"github.com/stretchr/testify/assert"
)

func TestWhenBuildingSuccessResult(t *testing.T) {
	t.Parallel()

	t.Run("Given an error-only use case succeeded", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return the success payload", func(t *testing.T) {
			t.Parallel()

			result := utils.SuccessResult()

			assert.Equal(t, map[string]any{"success": true}, result)
		})
	})
}
