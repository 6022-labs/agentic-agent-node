package utils_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
	"github.com/stretchr/testify/assert"
)

func TestWhenBuildingObjectProperty(t *testing.T) {
	t.Parallel()

	t.Run("Given a description", func(t *testing.T) {
		t.Parallel()

		t.Run("Should build a described object property", func(t *testing.T) {
			t.Parallel()

			property := utils.ObjectProperty("rule config")

			assert.Equal(t, "object", property.Type)
			if assert.NotNil(t, property.Description) {
				assert.Equal(t, "rule config", *property.Description)
			}
			// Free-form objects must allow extra keys or strict providers reject them.
			if assert.NotNil(t, property.AdditionalProperties) {
				assert.True(t, *property.AdditionalProperties)
			}
		})
	})
}
