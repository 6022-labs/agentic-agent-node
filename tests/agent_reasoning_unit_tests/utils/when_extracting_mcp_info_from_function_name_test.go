package utils_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/utils"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
)

func TestWhenExtractingMcpInfoFromFunctionName(t *testing.T) {
	t.Parallel()

	t.Run("Given a function name outside the mcp namespace", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail naming the namespace", func(t *testing.T) {
			t.Parallel()

			_, _, err := utils.ExtractMcpInfoFromFunctionName("self_get_registration", nil)

			assert.ErrorContains(t, err, "outside the mcp namespace")
		})
	})

	t.Run("Given the connection or tool does not exist", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail", func(t *testing.T) {
			t.Parallel()

			mcpConnection := model_builders.NewMCPConnectionBuilder().WithName("inventory").Build()
			mcpTools := map[*models.MCPConnection][]*models.Tool{mcpConnection: {{Name: "fetch_items"}}}

			_, _, err := utils.ExtractMcpInfoFromFunctionName("mcp_ghost_tool", mcpTools)
			assert.ErrorContains(t, err, "no MCP connection found")

			_, _, err = utils.ExtractMcpInfoFromFunctionName("mcp_inventory_ghost", mcpTools)
			assert.ErrorContains(t, err, "no MCP tool found")
		})
	})

	t.Run("Given a namespaced function name", func(t *testing.T) {
		t.Parallel()

		t.Run("Should resolve the connection name and the tool name", func(t *testing.T) {
			t.Parallel()

			mcpConnection := model_builders.NewMCPConnectionBuilder().WithName("inventory").Build()
			mcpTool := &models.Tool{Name: "fetch_items"}
			mcpTools := map[*models.MCPConnection][]*models.Tool{mcpConnection: {mcpTool}}

			conn, tool, err := utils.ExtractMcpInfoFromFunctionName("mcp_inventory_fetch_items", mcpTools)

			assert.NoError(t, err)
			assert.Equal(t, mcpConnection, conn)
			assert.Equal(t, mcpTool, tool)
		})
	})
}
