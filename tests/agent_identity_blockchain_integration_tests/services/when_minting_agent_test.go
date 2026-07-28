package services_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/services"
	"github.com/stretchr/testify/assert"
)

type WhenMintingAgentTestingSuite struct {
	fixture *chainFixture
	sut     *services.AgentMinter
}

func WhenMintingAgentBeforeEach(t *testing.T) *WhenMintingAgentTestingSuite {
	fixture := newChainFixture(t)
	return &WhenMintingAgentTestingSuite{
		fixture: fixture,
		sut:     fixture.agentMinter(),
	}
}

func TestWhenMintingAgent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given the node wallet is a moderator of the collection", func(t *testing.T) {
		t.Parallel()

		t.Run("Should mint the agent and make it resolvable on chain", func(t *testing.T) {
			t.Parallel()
			suite := WhenMintingAgentBeforeEach(t)

			txHash, err := suite.sut.Mint(ctx, suite.fixture.chainId(), mintParams(suite.fixture, "minted-agent"))

			assert.NoError(t, err)
			assert.NotEmpty(t, txHash)

			collectionRequester := suite.fixture.agentCollectionRequester()
			tokenId, err := collectionRequester.NameToTokenId(
				suite.fixture.chainId(),
				suite.fixture.stack.Collection,
				"minted-agent",
			)
			assert.NoError(t, err)

			owner, err := collectionRequester.OwnerOf(suite.fixture.chainId(), suite.fixture.stack.Collection, tokenId)
			assert.NoError(t, err)
			assert.Equal(t, suite.fixture.anvilAccountSettings.Address, owner)
		})

		t.Run("Should reject a second agent reusing the same name", func(t *testing.T) {
			t.Parallel()
			suite := WhenMintingAgentBeforeEach(t)

			_, err := suite.sut.Mint(ctx, suite.fixture.chainId(), mintParams(suite.fixture, "duplicate-agent"))
			assert.NoError(t, err)

			_, err = suite.sut.Mint(ctx, suite.fixture.chainId(), mintParams(suite.fixture, "duplicate-agent"))
			assert.Error(t, err)
		})
	})

	t.Run("Given a moderated collection taking mint proposals", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return the proposal id of each created proposal", func(t *testing.T) {
			t.Parallel()
			suite := WhenMintingAgentBeforeEach(t)

			txHash, firstProposalId, err := suite.sut.CreateMintProposal(
				ctx,
				suite.fixture.chainId(),
				mintParams(suite.fixture, "proposed-agent"),
			)
			assert.NoError(t, err)
			assert.NotEmpty(t, txHash)

			_, secondProposalId, err := suite.sut.CreateMintProposal(
				ctx,
				suite.fixture.chainId(),
				mintParams(suite.fixture, "second-proposed-agent"),
			)
			assert.NoError(t, err)
			assert.Equal(t, firstProposalId+1, secondProposalId)
		})
	})
}
