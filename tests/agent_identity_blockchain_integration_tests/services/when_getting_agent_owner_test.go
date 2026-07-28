package services_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/services"
	"github.com/stretchr/testify/assert"
)

type WhenGettingAgentOwnerTestingSuite struct {
	fixture *chainFixture
	sut     *services.AgentCollectionRequester
}

func WhenGettingAgentOwnerBeforeEach(t *testing.T) *WhenGettingAgentOwnerTestingSuite {
	fixture := newChainFixture(t)
	return &WhenGettingAgentOwnerTestingSuite{
		fixture: fixture,
		sut:     fixture.agentCollectionRequester(),
	}
}

func TestWhenGettingAgentOwner(t *testing.T) {
	t.Parallel()

	t.Run("Given a minted agent", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return the minting wallet as owner", func(t *testing.T) {
			t.Parallel()
			suite := WhenGettingAgentOwnerBeforeEach(t)
			tokenId := suite.fixture.mintAgent(t, "owned-agent")

			owner, err := suite.sut.OwnerOf(suite.fixture.chainId(), suite.fixture.stack.Collection, tokenId)

			assert.NoError(t, err)
			assert.Equal(t, suite.fixture.anvilAccountSettings.Address, owner)
		})
	})

	t.Run("Given a token id that was never minted", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail rather than report a zero owner", func(t *testing.T) {
			t.Parallel()
			suite := WhenGettingAgentOwnerBeforeEach(t)

			_, err := suite.sut.OwnerOf(suite.fixture.chainId(), suite.fixture.stack.Collection, 999)

			assert.Error(t, err)
		})
	})
}
