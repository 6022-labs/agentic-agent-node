package services_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/services"
	"github.com/stretchr/testify/assert"
)

type WhenCreatingMintProposalTestingSuite struct {
	fixture *chainFixture
	sut     *services.AgentMinter
}

func WhenCreatingMintProposalBeforeEach(t *testing.T) *WhenCreatingMintProposalTestingSuite {
	fixture := newChainFixture(t)
	return &WhenCreatingMintProposalTestingSuite{
		fixture: fixture,
		sut:     fixture.agentMinter(),
	}
}

func TestWhenCreatingMintProposal(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given a moderated collection taking mint proposals", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return the transaction hash", func(t *testing.T) {
			t.Parallel()
			suite := WhenCreatingMintProposalBeforeEach(t)

			txHash, _, err := suite.sut.CreateMintProposal(
				ctx,
				suite.fixture.chainId(),
				mintParams(suite.fixture, "proposed-agent"),
			)

			assert.NoError(t, err)
			assert.NotEmpty(t, txHash)
		})

		t.Run("Should increment the proposal id on each proposal", func(t *testing.T) {
			t.Parallel()
			suite := WhenCreatingMintProposalBeforeEach(t)
			_, firstProposalId, err := suite.sut.CreateMintProposal(
				ctx,
				suite.fixture.chainId(),
				mintParams(suite.fixture, "proposed-agent"),
			)
			assert.NoError(t, err)

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
