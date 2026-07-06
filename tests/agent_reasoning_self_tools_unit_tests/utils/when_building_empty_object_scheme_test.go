package utils_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
	"github.com/stretchr/testify/assert"
)

func TestWhenBuildingEmptyObjectScheme(t *testing.T) {
	t.Parallel()

	t.Run("Given a tool without arguments", func(t *testing.T) {
		t.Parallel()

		t.Run("Should build an object scheme with no properties", func(t *testing.T) {
			t.Parallel()

			scheme := utils.EmptyObjectScheme()

			assert.Equal(t, "object", scheme.Type)
			assert.Empty(t, scheme.Properties)
			assert.Empty(t, scheme.Required)
		})
	})
}
