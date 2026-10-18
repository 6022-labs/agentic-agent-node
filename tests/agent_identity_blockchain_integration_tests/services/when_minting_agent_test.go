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

		t.Run("Should return the transaction hash", func(t *testing.T) {
			t.Parallel()
			suite := WhenMintingAgentBeforeEach(t)

			txHash, err := suite.sut.Mint(ctx, suite.fixture.chainId(), mintParams(suite.fixture, "minted-agent"))

			assert.NoError(t, err)
			assert.NotEmpty(t, txHash)
		})

		t.Run("Should make the agent resolvable by name on chain", func(t *testing.T) {
			t.Parallel()
			suite := WhenMintingAgentBeforeEach(t)
			_, err := suite.sut.Mint(ctx, suite.fixture.chainId(), mintParams(suite.fixture, "minted-agent"))
			assert.NoError(t, err)

			_, err = suite.fixture.agentCollectionRequester().NameToTokenId(
				suite.fixture.chainId(),
				suite.fixture.stack.Collection,
				"minted-agent",
			)

			assert.NoError(t, err)
		})

		t.Run("Should make the minting wallet its owner", func(t *testing.T) {
			t.Parallel()
			suite := WhenMintingAgentBeforeEach(t)
			_, err := suite.sut.Mint(ctx, suite.fixture.chainId(), mintParams(suite.fixture, "minted-agent"))
			assert.NoError(t, err)
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
			assert.ErrorContains(t, err, "custom error 0x3fdf29d7")
		})
	})
}
