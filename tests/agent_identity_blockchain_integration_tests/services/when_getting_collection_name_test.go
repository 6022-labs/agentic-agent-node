package services_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/tests/agent_identity_blockchain_integration_tests/fixtures"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/services"
	"github.com/stretchr/testify/assert"
)

type WhenGettingCollectionNameTestingSuite struct {
	fixture *chainFixture
	sut     *services.AgentCollectionRequester
}

func WhenGettingCollectionNameBeforeEach(t *testing.T) *WhenGettingCollectionNameTestingSuite {
	fixture := newChainFixture(t)
	return &WhenGettingCollectionNameTestingSuite{
		fixture: fixture,
		sut:     fixture.agentCollectionRequester(),
	}
}

func TestWhenGettingCollectionName(t *testing.T) {
	t.Parallel()

	t.Run("Given a freshly created collection", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report the name it was created with", func(t *testing.T) {
			t.Parallel()
			suite := WhenGettingCollectionNameBeforeEach(t)

			name, err := suite.sut.CollectionName(suite.fixture.chainId(), suite.fixture.stack.Collection)

			assert.NoError(t, err)
			assert.Equal(t, fixtures.StackCollectionName, name)
		})
	})
}
