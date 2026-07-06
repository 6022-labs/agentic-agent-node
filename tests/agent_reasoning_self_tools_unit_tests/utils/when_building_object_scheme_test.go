package utils_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
	"github.com/stretchr/testify/assert"
)

func TestWhenBuildingObjectScheme(t *testing.T) {
	t.Parallel()

	t.Run("Given properties and required keys", func(t *testing.T) {
		t.Parallel()

		t.Run("Should carry them into an object scheme", func(t *testing.T) {
			t.Parallel()

			properties := map[string]models.ToolInputSchemeProperty{
				"chainId": utils.StringProperty("the chain"),
			}

			scheme := utils.ObjectScheme(properties, []string{"chainId"})

			assert.Equal(t, "object", scheme.Type)
			assert.Equal(t, properties, scheme.Properties)
			assert.Equal(t, []string{"chainId"}, scheme.Required)
		})
	})
}
