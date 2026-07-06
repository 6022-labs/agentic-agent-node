package utils_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_inference/utils"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
)

func TestWhenMappingProviderExecutedToolFunctionName(t *testing.T) {
	t.Parallel()

	t.Run("Given the echoed server label matches a known connection", func(t *testing.T) {
		t.Parallel()

		t.Run("Should rebuild the local mcp namespaced name", func(t *testing.T) {
			t.Parallel()

			mcpConnection := model_builders.NewMCPConnectionBuilder().WithName("weather").Build()
			mcpTools := map[*models.MCPConnection][]*models.Tool{mcpConnection: nil}

			name := utils.ProviderExecutedToolFunctionName(mcpTools, "weather", "get")

			assert.Equal(t, "mcp_weather_get", name)
		})
	})

	t.Run("Given the echoed server label matches no connection", func(t *testing.T) {
		t.Parallel()

		t.Run("Should keep the provider's raw tool name", func(t *testing.T) {
			t.Parallel()

			mcpConnection := model_builders.NewMCPConnectionBuilder().WithName("weather").Build()
			mcpTools := map[*models.MCPConnection][]*models.Tool{mcpConnection: nil}

			name := utils.ProviderExecutedToolFunctionName(mcpTools, "normalized-weather", "get")

			assert.Equal(t, "get", name)
		})
	})
}
