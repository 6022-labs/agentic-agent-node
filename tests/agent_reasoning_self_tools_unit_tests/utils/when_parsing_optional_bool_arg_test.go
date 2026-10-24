package utils_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestWhenParsingOptionalBoolArg(t *testing.T) {
	t.Parallel()

	t.Run("Given a missing or unparseable value", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return nil", func(t *testing.T) {
			t.Parallel()

			assert.Nil(t, utils.OptionalBoolArg(nil, "background"))
			assert.Nil(t, utils.OptionalBoolArg(map[string]any{"background": nil}, "background"))
			assert.Nil(t, utils.OptionalBoolArg(map[string]any{"background": "yes please"}, "background"))
		})
	})

	t.Run("Given a truthy value in any json shape", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return true", func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, lo.ToPtr(true), utils.OptionalBoolArg(map[string]any{"background": true}, "background"))
			assert.Equal(t, lo.ToPtr(true), utils.OptionalBoolArg(map[string]any{"background": "true"}, "background"))
			assert.Equal(
				t,
				lo.ToPtr(true),
				utils.OptionalBoolArg(map[string]any{"background": float64(1)}, "background"),
			)
		})
	})

	t.Run("Given an explicit false", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return it rather than nil, so the caller can tell it apart from an omission", func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, lo.ToPtr(false), utils.OptionalBoolArg(map[string]any{"background": false}, "background"))
			assert.Equal(t, lo.ToPtr(false), utils.OptionalBoolArg(map[string]any{"background": "false"}, "background"))
		})
	})
}
