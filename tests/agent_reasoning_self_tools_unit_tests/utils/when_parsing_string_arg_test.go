package utils_test

import (
	"encoding/json"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
	"github.com/stretchr/testify/assert"
)

func TestWhenParsingStringArg(t *testing.T) {
	t.Parallel()

	t.Run("Given the argument is present", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return its value", func(t *testing.T) {
			t.Parallel()

			value, err := utils.StringArg(map[string]any{"name": "weather"}, "name")

			assert.NoError(t, err)
			assert.Equal(t, "weather", value)
		})
	})

	t.Run("Given the argument is present but empty", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return the empty string, not an error", func(t *testing.T) {
			t.Parallel()

			value, err := utils.StringArg(map[string]any{"content": ""}, "content")

			assert.NoError(t, err)
			assert.Empty(t, value)
		})
	})

	t.Run("Given the argument is a numeric or bool value", func(t *testing.T) {
		t.Parallel()

		t.Run("Should convert it to a string", func(t *testing.T) {
			t.Parallel()

			value, err := utils.StringArg(map[string]any{"chainId": float64(80002)}, "chainId")

			assert.NoError(t, err)
			assert.Equal(t, "80002", value)
		})

		t.Run("Should render large numbers in plain decimal, never exponent form", func(t *testing.T) {
			t.Parallel()

			for arg, expected := range map[float64]string{
				1000000:  "1000000",  // fmt.Sprint would yield "1e+06"
				11155111: "11155111", // Sepolia chain id — "1.1155111e+07"
			} {
				value, err := utils.StringArg(map[string]any{"price": arg}, "price")

				assert.NoError(t, err)
				assert.Equal(t, expected, value)
			}
		})

		t.Run("Should accept a json.Number without losing precision", func(t *testing.T) {
			t.Parallel()

			value, err := utils.StringArg(map[string]any{"price": json.Number("1000000000000000000")}, "price")

			assert.NoError(t, err)
			assert.Equal(t, "1000000000000000000", value)
		})
	})

	t.Run("Given the argument is missing", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail naming the argument", func(t *testing.T) {
			t.Parallel()

			for _, args := range []map[string]any{nil, {}, {"name": nil}} {
				value, err := utils.StringArg(args, "name")

				assert.Empty(t, value)
				assert.ErrorContains(t, err, "name")
			}
		})
	})

	t.Run("Given the argument is a non-scalar type", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail reporting a type mismatch", func(t *testing.T) {
			t.Parallel()

			value, err := utils.StringArg(map[string]any{"name": map[string]any{}}, "name")

			assert.Empty(t, value)
			assert.ErrorContains(t, err, "must be a string")
		})
	})
}
