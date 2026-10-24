package utils_test

import (
	"errors"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
	"github.com/stretchr/testify/assert"
)

func TestWhenConvertingToToolResult(t *testing.T) {
	t.Parallel()

	t.Run("Given the use case fails", func(t *testing.T) {
		t.Parallel()

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			out, err := utils.ToToolResult(nil, errors.New("boom"))

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "boom")
		})
	})

	t.Run("Given the response cannot be encoded", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail", func(t *testing.T) {
			t.Parallel()

			out, err := utils.ToToolResult(make(chan int), nil)

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "failed to encode tool result")
		})
	})

	t.Run("Given a use-case response", func(t *testing.T) {
		t.Parallel()

		t.Run("Should encode it as a generic map", func(t *testing.T) {
			t.Parallel()

			response := struct {
				Name string `json:"name"`
			}{Name: "agent-x"}

			out, err := utils.ToToolResult(&response, nil)

			assert.NoError(t, err)
			assert.Equal(t, map[string]any{"name": "agent-x"}, out)
		})
	})
}
