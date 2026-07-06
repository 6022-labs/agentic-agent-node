package utils_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/utils"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
)

func TestWhenMappingMcpToolToFunctionName(t *testing.T) {
	t.Parallel()

	t.Run("Given an mcp connection and one of its tools", func(t *testing.T) {
		t.Parallel()

		t.Run("Should prefix the function name with the mcp namespace and connection name", func(t *testing.T) {
			t.Parallel()

			mcpConnection := model_builders.NewMCPConnectionBuilder().WithName("inventory").Build()

			name := utils.McpToolToFunctionName(mcpConnection, &models.Tool{Name: "fetch_items"})

			assert.Equal(t, "mcp_inventory_fetch_items", name)
		})
	})
}
