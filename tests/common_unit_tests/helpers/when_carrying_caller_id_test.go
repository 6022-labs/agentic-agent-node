package helpers_test

import (
	"context"
	"strings"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/common/helpers"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/stretchr/testify/assert"
)

func TestWhenCarryingCallerId(t *testing.T) {
	t.Parallel()

	t.Run("Given no caller id", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report no caller", func(t *testing.T) {
			t.Parallel()

			_, ok := helpers.CallerIdFromContext(context.Background())

			assert.False(t, ok)
		})
	})

	t.Run("Given a blank caller id", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report no caller", func(t *testing.T) {
			t.Parallel()

			ctx := helpers.ContextWithCallerId(context.Background(), "   ")

			_, ok := helpers.CallerIdFromContext(ctx)

			assert.False(t, ok)
		})
	})

	t.Run("Given a padded mixed-case caller id", func(t *testing.T) {
		t.Parallel()

		t.Run("Should read it back trimmed and lowercased", func(t *testing.T) {
			t.Parallel()

			callerId := strings.ToLower(common_tests.RandomEthAddress().Hex())
			ctx := helpers.ContextWithCallerId(context.Background(), " "+strings.ToUpper(callerId)+" ")

			carried, ok := helpers.CallerIdFromContext(ctx)

			assert.True(t, ok)
			assert.Equal(t, callerId, carried)
		})
	})
}
