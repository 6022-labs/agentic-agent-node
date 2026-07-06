package utils_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
	"github.com/stretchr/testify/assert"
)

func TestWhenBuildingStringProperty(t *testing.T) {
	t.Parallel()

	t.Run("Given a description", func(t *testing.T) {
		t.Parallel()

		t.Run("Should build a described string property", func(t *testing.T) {
			t.Parallel()

			property := utils.StringProperty("the chain")

			assert.Equal(t, "string", property.Type)
			if assert.NotNil(t, property.Description) {
				assert.Equal(t, "the chain", *property.Description)
			}
		})
	})
}
