package utils_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
	"github.com/stretchr/testify/assert"
)

func TestWhenParsingJsonArrayArg(t *testing.T) {
	t.Parallel()

	t.Run("Given the argument is missing", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail naming the argument", func(t *testing.T) {
			t.Parallel()

			value, err := utils.JsonArrayArg[string](map[string]any{}, "payToPatterns")

			assert.Nil(t, value)
			assert.ErrorContains(t, err, `argument "payToPatterns" is required`)
		})
	})

	t.Run("Given an empty array", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return an empty slice", func(t *testing.T) {
			t.Parallel()

			value, err := utils.JsonArrayArg[string](map[string]any{"payToPatterns": []any{}}, "payToPatterns")

			assert.NoError(t, err)
			assert.Equal(t, []string{}, value)
		})
	})

	t.Run("Given an array of strings", func(t *testing.T) {
		t.Parallel()

		t.Run("Should decode every item", func(t *testing.T) {
			t.Parallel()

			value, err := utils.JsonArrayArg[string](map[string]any{"payToPatterns": []any{".*"}}, "payToPatterns")

			assert.NoError(t, err)
			assert.Equal(t, []string{".*"}, value)
		})
	})
}
