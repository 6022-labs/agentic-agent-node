package utils_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/utils"
	"github.com/stretchr/testify/assert"
)

func TestWhenBuildingFunctionName(t *testing.T) {
	t.Parallel()

	t.Run("Given a namespace and a tool name", func(t *testing.T) {
		t.Parallel()

		t.Run("Should join them with the separator", func(t *testing.T) {
			t.Parallel()

			name := utils.ToFunctionName("self", "get_registration")

			assert.Equal(t, "self_get_registration", name)
		})
	})
}
