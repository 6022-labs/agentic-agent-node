package utils_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/utils"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
)

func TestWhenFindingMcpConnectionByName(t *testing.T) {
	t.Parallel()

	t.Run("Given no connection with that name", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report not found", func(t *testing.T) {
			t.Parallel()

			inventory := model_builders.NewMCPConnectionBuilder().WithName("inventory").Build()
			mcpTools := map[*models.MCPConnection][]*models.Tool{inventory: nil}

			found, ok := utils.FindMcpConnectionByName(mcpTools, "weather")

			assert.False(t, ok)
			assert.Nil(t, found)
		})
	})

	t.Run("Given a connection with that name exists", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return it", func(t *testing.T) {
			t.Parallel()

			weather := model_builders.NewMCPConnectionBuilder().WithName("weather").Build()
			inventory := model_builders.NewMCPConnectionBuilder().WithName("inventory").Build()
			mcpTools := map[*models.MCPConnection][]*models.Tool{weather: nil, inventory: nil}

			found, ok := utils.FindMcpConnectionByName(mcpTools, "weather")

			assert.True(t, ok)
			assert.Equal(t, weather, found)
		})
	})
}
