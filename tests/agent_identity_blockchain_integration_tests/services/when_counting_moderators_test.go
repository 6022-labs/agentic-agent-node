package services_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/services"
	"github.com/stretchr/testify/assert"
)

type WhenCountingModeratorsTestingSuite struct {
	fixture *chainFixture
	sut     *services.AgentCollectionRequester
}

func WhenCountingModeratorsBeforeEach(t *testing.T) *WhenCountingModeratorsTestingSuite {
	fixture := newChainFixture(t)
	return &WhenCountingModeratorsTestingSuite{
		fixture: fixture,
		sut:     fixture.agentCollectionRequester(),
	}
}

func TestWhenCountingModerators(t *testing.T) {
	t.Parallel()

	t.Run("Given a freshly created collection", func(t *testing.T) {
		t.Parallel()

		t.Run("Should count the single moderator it was created with", func(t *testing.T) {
			t.Parallel()
			suite := WhenCountingModeratorsBeforeEach(t)

			count, err := suite.sut.ModeratorCount(suite.fixture.chainId(), suite.fixture.stack.Collection)

			assert.NoError(t, err)
			assert.Equal(t, uint64(1), count)
		})
	})
}
