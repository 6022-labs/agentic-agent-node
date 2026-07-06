package utils_test

import (
	"encoding/json"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
	"github.com/stretchr/testify/assert"
)

func TestWhenParsingRawJsonArg(t *testing.T) {
	t.Parallel()

	t.Run("Given an object argument", func(t *testing.T) {
		t.Parallel()

		t.Run("Should re-encode it as raw json", func(t *testing.T) {
			t.Parallel()

			raw, err := utils.RawJsonArg(map[string]any{"config": map[string]any{"network": "eip155:80002"}}, "config")

			assert.NoError(t, err)
			assert.Equal(t, json.RawMessage(`{"network":"eip155:80002"}`), raw)
		})
	})

	t.Run("Given the argument is missing or nil", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail naming the argument", func(t *testing.T) {
			t.Parallel()

			for _, args := range []map[string]any{nil, {}, {"config": nil}} {
				raw, err := utils.RawJsonArg(args, "config")

				assert.Nil(t, raw)
				assert.ErrorContains(t, err, "config")
			}
		})
	})

	t.Run("Given the argument cannot be encoded", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail", func(t *testing.T) {
			t.Parallel()

			raw, err := utils.RawJsonArg(map[string]any{"config": make(chan int)}, "config")

			assert.Nil(t, raw)
			assert.ErrorContains(t, err, "config")
		})
	})
}
