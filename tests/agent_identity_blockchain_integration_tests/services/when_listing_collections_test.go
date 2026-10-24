package services_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	"github.com/6022-labs/agentic-agent-node/tests/agent_identity_blockchain_integration_tests/fixtures"
	"github.com/ethereum/go-ethereum/common"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/services"
	"github.com/stretchr/testify/assert"
)

type WhenListingCollectionsTestingSuite struct {
	fixture *chainFixture
	sut     *services.AgentCollectionsManagerRequester
}

func WhenListingCollectionsBeforeEach(t *testing.T) *WhenListingCollectionsTestingSuite {
	fixture := newChainFixture(t)
	return &WhenListingCollectionsTestingSuite{
		fixture: fixture,
		sut:     fixture.agentCollectionsManagerRequester(),
	}
}

func TestWhenListingCollections(t *testing.T) {
	t.Parallel()

	t.Run("Given a chain with no configured manager", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail instead of calling a zero address", func(t *testing.T) {
			t.Parallel()
			fixture := newChainFixture(t)
			emptySettings := settings.AgentCollectionsManagersSettings{}
			sut := services.NewAgentCollectionsManagerRequester(fixture.ethClientFactory, &emptySettings)

			_, err := sut.ListCollections(fixture.chainId())

			assert.ErrorContains(t, err, "no AgentCollectionsManager configured for chainId")
		})
	})

	t.Run("Given a manager governing one created collection", func(t *testing.T) {
		t.Parallel()

		t.Run("Should list its address", func(t *testing.T) {
			t.Parallel()
			suite := WhenListingCollectionsBeforeEach(t)

			collections, err := suite.sut.ListCollections(suite.fixture.chainId())

			assert.NoError(t, err)
			if assert.Len(t, collections, 1) {
				assert.Equal(t, suite.fixture.stack.Collection.Hex(), common.HexToAddress(collections[0].Address).Hex())
			}
		})

		t.Run("Should list its name", func(t *testing.T) {
			t.Parallel()
			suite := WhenListingCollectionsBeforeEach(t)

			collections, err := suite.sut.ListCollections(suite.fixture.chainId())

			assert.NoError(t, err)
			if assert.Len(t, collections, 1) {
				assert.Equal(t, fixtures.StackCollectionName, collections[0].Name)
			}
		})

		t.Run("Should list its description", func(t *testing.T) {
			t.Parallel()
			suite := WhenListingCollectionsBeforeEach(t)

			collections, err := suite.sut.ListCollections(suite.fixture.chainId())

			assert.NoError(t, err)
			if assert.Len(t, collections, 1) {
				assert.Equal(t, "integration test collection", collections[0].Description)
			}
		})

		t.Run("Should list its moderator count", func(t *testing.T) {
			t.Parallel()
			suite := WhenListingCollectionsBeforeEach(t)

			collections, err := suite.sut.ListCollections(suite.fixture.chainId())

			assert.NoError(t, err)
			if assert.Len(t, collections, 1) {
				assert.Equal(t, uint64(1), collections[0].ModeratorCount)
			}
		})
	})
}
