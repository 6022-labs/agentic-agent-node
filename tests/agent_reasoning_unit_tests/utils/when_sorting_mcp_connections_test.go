package utils_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/utils"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
)

func TestWhenSortingMcpConnections(t *testing.T) {
	t.Parallel()

	t.Run("Given tools keyed by connection", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return the connections in name order", func(t *testing.T) {
			t.Parallel()

			weather := model_builders.NewMCPConnectionBuilder().WithName("weather").Build()
			inventory := model_builders.NewMCPConnectionBuilder().WithName("inventory").Build()
			mcpTools := map[*models.MCPConnection][]*models.Tool{
				weather:   nil,
				inventory: nil,
			}

			sorted := utils.SortedMcpConnections(mcpTools)

			assert.Equal(t, []*models.MCPConnection{inventory, weather}, sorted)
		})
	})

	t.Run("Given no connections", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return an empty slice", func(t *testing.T) {
			t.Parallel()

			sorted := utils.SortedMcpConnections(nil)

			assert.Empty(t, sorted)
		})
	})
}
