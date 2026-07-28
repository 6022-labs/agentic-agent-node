package services_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/services"
	"github.com/stretchr/testify/assert"
)

type WhenResolvingTokenIdFromNameTestingSuite struct {
	fixture *chainFixture
	sut     *services.AgentCollectionRequester
}

func WhenResolvingTokenIdFromNameBeforeEach(t *testing.T) *WhenResolvingTokenIdFromNameTestingSuite {
	fixture := newChainFixture(t)
	return &WhenResolvingTokenIdFromNameTestingSuite{
		fixture: fixture,
		sut:     fixture.agentCollectionRequester(),
	}
}

func TestWhenResolvingTokenIdFromName(t *testing.T) {
	t.Parallel()

	t.Run("Given a minted agent", func(t *testing.T) {
		t.Parallel()

		agentName := "named-agent"

		t.Run("Should resolve its token id", func(t *testing.T) {
			t.Parallel()
			suite := WhenResolvingTokenIdFromNameBeforeEach(t)
			tokenId := suite.fixture.mintAgent(t, agentName)

			resolved, err := suite.sut.NameToTokenId(suite.fixture.chainId(), suite.fixture.stack.Collection, agentName)

			assert.NoError(t, err)
			assert.Equal(t, tokenId, resolved)
		})
	})
}
