package utils_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
	"github.com/stretchr/testify/assert"
)

func TestWhenBuildingArrayProperty(t *testing.T) {
	t.Parallel()

	t.Run("Given a description and an item property", func(t *testing.T) {
		t.Parallel()

		t.Run("Should build a described array of that item", func(t *testing.T) {
			t.Parallel()

			property := utils.ArrayProperty("payees", utils.StringProperty("one payee"))

			assert.Equal(t, "array", property.Type)
			if assert.NotNil(t, property.Description) {
				assert.Equal(t, "payees", *property.Description)
			}
			if assert.NotNil(t, property.Items) {
				assert.Equal(t, "string", property.Items.Type)
			}
		})
	})
}
