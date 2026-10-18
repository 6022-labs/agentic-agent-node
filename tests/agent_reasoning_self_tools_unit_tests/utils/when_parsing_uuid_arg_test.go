package utils_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestWhenParsingUuidArg(t *testing.T) {
	t.Parallel()

	t.Run("Given a valid uuid argument", func(t *testing.T) {
		t.Parallel()

		t.Run("Should parse it", func(t *testing.T) {
			t.Parallel()

			id := uuid.New()

			parsed, err := utils.UuidArg(map[string]any{"id": id.String()}, "id")

			assert.NoError(t, err)
			assert.Equal(t, id, parsed)
		})
	})

	t.Run("Given a malformed argument", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail", func(t *testing.T) {
			t.Parallel()

			parsed, err := utils.UuidArg(map[string]any{"id": "nope"}, "id")

			assert.Equal(t, uuid.Nil, parsed)
			assert.ErrorContains(t, err, "uuid")
		})
	})

	t.Run("Given the argument is missing", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail naming the argument", func(t *testing.T) {
			t.Parallel()

			parsed, err := utils.UuidArg(nil, "id")

			assert.Equal(t, uuid.Nil, parsed)
			assert.ErrorContains(t, err, "id")
		})
	})
}
