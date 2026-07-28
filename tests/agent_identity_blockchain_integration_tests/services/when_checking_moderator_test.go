package services_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/services"
	"github.com/stretchr/testify/assert"
)

type WhenCheckingModeratorTestingSuite struct {
	fixture *chainFixture
	sut     *services.AgentCollectionRequester
}

func WhenCheckingModeratorBeforeEach(t *testing.T) *WhenCheckingModeratorTestingSuite {
	fixture := newChainFixture(t)
	return &WhenCheckingModeratorTestingSuite{
		fixture: fixture,
		sut:     fixture.agentCollectionRequester(),
	}
}

func TestWhenCheckingModerator(t *testing.T) {
	t.Parallel()

	t.Run("Given a freshly created collection", func(t *testing.T) {
		t.Parallel()

		t.Run("Should recognise the moderator it was created with", func(t *testing.T) {
			t.Parallel()
			suite := WhenCheckingModeratorBeforeEach(t)

			isModerator, err := suite.sut.IsModerator(
				suite.fixture.chainId(),
				suite.fixture.stack.Collection,
				suite.fixture.anvilAccountSettings.Address,
			)

			assert.NoError(t, err)
			assert.True(t, isModerator)
		})

		t.Run("Should reject an unrelated account", func(t *testing.T) {
			t.Parallel()
			suite := WhenCheckingModeratorBeforeEach(t)

			isModerator, err := suite.sut.IsModerator(
				suite.fixture.chainId(),
				suite.fixture.stack.Collection,
				common.HexToAddress("0x000000000000000000000000000000000000dEaD"),
			)

			assert.NoError(t, err)
			assert.False(t, isModerator)
		})
	})
}
