package utils_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
	"github.com/stretchr/testify/assert"
)

type jsonObjectArgTarget struct {
	Network string `json:"network"`
}

func TestWhenParsingJsonObjectArg(t *testing.T) {
	t.Parallel()

	t.Run("Given the argument is missing", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail naming the argument", func(t *testing.T) {
			t.Parallel()

			value, err := utils.JsonObjectArg[jsonObjectArgTarget](map[string]any{}, "config")

			assert.Nil(t, value)
			assert.ErrorContains(t, err, "config")
		})
	})

	t.Run("Given an object that does not fit the target type", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail as the wrong shape", func(t *testing.T) {
			t.Parallel()

			value, err := utils.JsonObjectArg[jsonObjectArgTarget](
				map[string]any{"config": map[string]any{"network": 42.0}},
				"config",
			)

			assert.Nil(t, value)
			assert.ErrorContains(t, err, "has the wrong shape")
		})
	})

	t.Run("Given an object that fits the target type", func(t *testing.T) {
		t.Parallel()

		t.Run("Should decode it", func(t *testing.T) {
			t.Parallel()

			value, err := utils.JsonObjectArg[jsonObjectArgTarget](
				map[string]any{"config": map[string]any{"network": "eip155:80002"}},
				"config",
			)

			assert.NoError(t, err)
			if assert.NotNil(t, value) {
				assert.Equal(t, "eip155:80002", value.Network)
			}
		})
	})
}
