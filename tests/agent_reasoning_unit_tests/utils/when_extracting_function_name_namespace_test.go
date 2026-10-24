package utils_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/utils"
	"github.com/stretchr/testify/assert"
)

func TestWhenExtractingFunctionNameNamespace(t *testing.T) {
	t.Parallel()

	t.Run("Given a function name without a separator", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return an empty namespace", func(t *testing.T) {
			t.Parallel()

			namespace := utils.ExtractNamespaceFromFunctionName("standalone")

			assert.Empty(t, namespace)
		})
	})

	t.Run("Given a namespaced function name", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return its namespace segment", func(t *testing.T) {
			t.Parallel()

			for functionName, expected := range map[string]string{
				"self_get_registration":     "self",
				"mcp_inventory_fetch_items": "mcp",
				"ghost_tool":                "ghost",
			} {
				namespace := utils.ExtractNamespaceFromFunctionName(functionName)

				assert.Equal(t, expected, namespace, functionName)
			}
		})
	})
}
