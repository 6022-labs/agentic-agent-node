package utils_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/utils"
	"github.com/stretchr/testify/assert"
)

func TestWhenExtractingSelfToolNameFromFunctionName(t *testing.T) {
	t.Parallel()

	t.Run("Given a function name outside the self namespace", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail naming the namespace", func(t *testing.T) {
			t.Parallel()

			for _, name := range []string{"inventory_fetch", "self", "selfish_tool"} {
				_, err := utils.ExtractSelfToolNameFromFunctionName(name)

				assert.ErrorContains(t, err, "outside the self namespace", name)
			}
		})
	})

	t.Run("Given a self-namespaced function name without a tool", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail", func(t *testing.T) {
			t.Parallel()

			_, err := utils.ExtractSelfToolNameFromFunctionName("self_")

			assert.ErrorContains(t, err, "invalid function name format")
		})
	})

	t.Run("Given a namespaced function name", func(t *testing.T) {
		t.Parallel()

		t.Run("Should extract the bare tool name", func(t *testing.T) {
			t.Parallel()

			toolName, err := utils.ExtractSelfToolNameFromFunctionName("self_get_registration")

			assert.NoError(t, err)
			assert.Equal(t, "get_registration", toolName)
		})
	})
}
