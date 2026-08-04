package utils_test

import (
	"encoding/json"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestWhenParsingOptionalIntArg(t *testing.T) {
	t.Parallel()

	t.Run("Given a numeric value in any json shape", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return it as an int", func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, lo.ToPtr(30), utils.OptionalIntArg(map[string]any{"timeout": float64(30)}, "timeout"))
			assert.Equal(t, lo.ToPtr(30), utils.OptionalIntArg(map[string]any{"timeout": json.Number("30")}, "timeout"))
			assert.Equal(t, lo.ToPtr(30), utils.OptionalIntArg(map[string]any{"timeout": "30"}, "timeout"))
			assert.Equal(t, lo.ToPtr(30), utils.OptionalIntArg(map[string]any{"timeout": 30}, "timeout"))
		})

		t.Run("Should truncate a fractional value", func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, lo.ToPtr(30), utils.OptionalIntArg(map[string]any{"timeout": 30.9}, "timeout"))
		})
	})

	t.Run("Given a missing or unparseable value", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return nil so the use case applies its default", func(t *testing.T) {
			t.Parallel()

			assert.Nil(t, utils.OptionalIntArg(nil, "timeout"))
			assert.Nil(t, utils.OptionalIntArg(map[string]any{"timeout": nil}, "timeout"))
			assert.Nil(t, utils.OptionalIntArg(map[string]any{"timeout": "soon"}, "timeout"))
			assert.Nil(t, utils.OptionalIntArg(map[string]any{"timeout": []any{1}}, "timeout"))
		})
	})

	t.Run("Given an explicit zero", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return it rather than nil, so the caller can tell it apart from an omission", func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, lo.ToPtr(0), utils.OptionalIntArg(map[string]any{"timeout": float64(0)}, "timeout"))
			assert.Equal(t, lo.ToPtr(0), utils.OptionalIntArg(map[string]any{"timeout": "0"}, "timeout"))
		})
	})
}
