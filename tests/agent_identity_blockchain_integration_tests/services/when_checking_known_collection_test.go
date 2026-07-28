package services_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/services"
	"github.com/stretchr/testify/assert"
)

type WhenCheckingKnownCollectionTestingSuite struct {
	fixture *chainFixture
	sut     *services.AgentCollectionsManagerRequester
}

func WhenCheckingKnownCollectionBeforeEach(t *testing.T) *WhenCheckingKnownCollectionTestingSuite {
	fixture := newChainFixture(t)
	return &WhenCheckingKnownCollectionTestingSuite{
		fixture: fixture,
		sut:     fixture.agentCollectionsManagerRequester(),
	}
}

func TestWhenCheckingKnownCollection(t *testing.T) {
	t.Parallel()

	t.Run("Given a manager governing one created collection", func(t *testing.T) {
		t.Parallel()

		t.Run("Should recognise the created collection", func(t *testing.T) {
			t.Parallel()
			suite := WhenCheckingKnownCollectionBeforeEach(t)

			isKnown, err := suite.sut.IsKnownCollection(suite.fixture.chainId(), suite.fixture.stack.Collection)

			assert.NoError(t, err)
			assert.True(t, isKnown)
		})

		t.Run("Should reject an address it never created", func(t *testing.T) {
			t.Parallel()
			suite := WhenCheckingKnownCollectionBeforeEach(t)

			isKnown, err := suite.sut.IsKnownCollection(
				suite.fixture.chainId(),
				common.HexToAddress("0x000000000000000000000000000000000000dEaD"),
			)

			assert.NoError(t, err)
			assert.False(t, isKnown)
		})
	})
}
