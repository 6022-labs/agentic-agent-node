package observability_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/common/observability"
	"github.com/stretchr/testify/assert"
)

func TestWhenCarryingAgentIdentityInContext(t *testing.T) {
	t.Parallel()

	t.Run("Given a bare context", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report no identity", func(t *testing.T) {
			t.Parallel()

			_, ok := observability.AgentIdentityFromContext(context.Background())

			assert.False(t, ok)
		})
	})

	t.Run("Given an identity has been attached", func(t *testing.T) {
		t.Parallel()

		t.Run("Should read it back from the derived context", func(t *testing.T) {
			t.Parallel()

			ctx := observability.ContextWithAgentIdentity(context.Background(), observability.AgentIdentity{EnsDomain: "myagent.6022.eth"})

			identity, ok := observability.AgentIdentityFromContext(ctx)

			assert.True(t, ok)
			assert.Equal(t, "myagent.6022.eth", identity.EnsDomain)
		})
	})
}
