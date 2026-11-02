package utils_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
	"github.com/stretchr/testify/assert"
)

type jsonArrayArgTarget struct {
	WindowSeconds int `json:"windowSeconds"`
}

func TestWhenParsingOptionalJsonArrayArg(t *testing.T) {
	t.Parallel()

	t.Run("Given the argument is missing", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return nil without error", func(t *testing.T) {
			t.Parallel()

			value, err := utils.OptionalJsonArrayArg[jsonArrayArgTarget](map[string]any{}, "windows")

			assert.NoError(t, err)
			assert.Nil(t, value)
		})
	})

	t.Run("Given the argument is not an array", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail naming the argument", func(t *testing.T) {
			t.Parallel()

			value, err := utils.OptionalJsonArrayArg[jsonArrayArgTarget](
				map[string]any{"windows": "daily"},
				"windows",
			)

			assert.Nil(t, value)
			assert.ErrorContains(t, err, `argument "windows" must be a json array`)
		})
	})

	t.Run("Given an array that does not fit the target type", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail as the wrong shape", func(t *testing.T) {
			t.Parallel()

			value, err := utils.OptionalJsonArrayArg[jsonArrayArgTarget](
				map[string]any{"windows": []any{map[string]any{"windowSeconds": []any{}}}},
				"windows",
			)

			assert.Nil(t, value)
			assert.ErrorContains(t, err, "has the wrong shape")
		})
	})

	t.Run("Given an array that fits the target type", func(t *testing.T) {
		t.Parallel()

		t.Run("Should decode every item", func(t *testing.T) {
			t.Parallel()

			value, err := utils.OptionalJsonArrayArg[jsonArrayArgTarget](
				map[string]any{"windows": []any{
					map[string]any{"windowSeconds": 3600.0},
					map[string]any{"windowSeconds": 86400.0},
				}},
				"windows",
			)

			assert.NoError(t, err)
			assert.Equal(t, []jsonArrayArgTarget{{WindowSeconds: 3600}, {WindowSeconds: 86400}}, value)
		})
	})
}
