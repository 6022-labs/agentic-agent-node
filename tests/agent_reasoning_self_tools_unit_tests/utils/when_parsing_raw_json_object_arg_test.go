package utils_test

import (
	"encoding/json"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
	"github.com/stretchr/testify/assert"
)

func TestWhenParsingRawJsonObjectArg(t *testing.T) {
	t.Parallel()

	t.Run("Given an object argument", func(t *testing.T) {
		t.Parallel()

		t.Run("Should re-encode it as raw json", func(t *testing.T) {
			t.Parallel()

			raw, err := utils.RawJsonObjectArg(
				map[string]any{"config": map[string]any{"network": "eip155:80002"}},
				"config",
			)

			assert.NoError(t, err)
			assert.Equal(t, json.RawMessage(`{"network":"eip155:80002"}`), raw)
		})
	})

	t.Run("Given the argument is missing or nil", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail naming the argument", func(t *testing.T) {
			t.Parallel()

			for _, args := range []map[string]any{nil, {}, {"config": nil}} {
				raw, err := utils.RawJsonObjectArg(args, "config")

				assert.Nil(t, raw)
				assert.ErrorContains(t, err, "config")
			}
		})
	})

	t.Run("Given the argument is not an object", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject it as not a json object", func(t *testing.T) {
			t.Parallel()

			// A stringified object (a common model mistake) and other scalars are all rejected.
			for _, value := range []any{`{"network":"eip155:80002"}`, 42.0, []any{"a"}, true} {
				raw, err := utils.RawJsonObjectArg(map[string]any{"config": value}, "config")

				assert.Nil(t, raw)
				assert.ErrorContains(t, err, "must be a json object")
			}
		})
	})

	t.Run("Given the object holds an unencodable value", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail", func(t *testing.T) {
			t.Parallel()

			raw, err := utils.RawJsonObjectArg(
				map[string]any{"config": map[string]any{"bad": make(chan int)}},
				"config",
			)

			assert.Nil(t, raw)
			assert.ErrorContains(t, err, "config")
		})
	})
}
