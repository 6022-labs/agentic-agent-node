package services_test

import (
	"testing"

	agent_enums "github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	"github.com/ethereum/go-ethereum/common"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/services"
	"github.com/stretchr/testify/assert"
)

type WhenFindingAgentByAddressTestingSuite struct {
	fixture *chainFixture
	sut     *services.AgentCollectionsManagerRequester
}

func WhenFindingAgentByAddressBeforeEach(t *testing.T) *WhenFindingAgentByAddressTestingSuite {
	fixture := newChainFixture(t)
	return &WhenFindingAgentByAddressTestingSuite{
		fixture: fixture,
		sut:     fixture.agentCollectionsManagerRequester(),
	}
}

func TestWhenFindingAgentByAddress(t *testing.T) {
	t.Parallel()

	t.Run("Given an agent minted in a governed collection", func(t *testing.T) {
		t.Parallel()

		agentName := "managed-agent"

		t.Run("Should find it by its registered evm address", func(t *testing.T) {
			t.Parallel()
			suite := WhenFindingAgentByAddressBeforeEach(t)
			tokenId := suite.fixture.mintAgent(t, agentName)

			collection, foundTokenId, name, err := suite.sut.FindAgentByAddress(
				suite.fixture.chainId(),
				agent_enums.AddressTypeEvm,
				suite.fixture.anvilAccountSettings.Address.Hex(),
			)

			assert.NoError(t, err)
			assert.Equal(t, suite.fixture.stack.Collection, collection)
			assert.Equal(t, tokenId, foundTokenId)
			assert.Equal(t, agentName, name)
		})

		t.Run("Should return a zero collection for an address that owns no agent", func(t *testing.T) {
			t.Parallel()
			suite := WhenFindingAgentByAddressBeforeEach(t)

			collection, tokenId, name, err := suite.sut.FindAgentByAddress(
				suite.fixture.chainId(),
				agent_enums.AddressTypeEvm,
				"0x000000000000000000000000000000000000dEaD",
			)

			assert.NoError(t, err)
			assert.Equal(t, common.Address{}, collection)
			assert.Zero(t, tokenId)
			assert.Empty(t, name)
		})
	})
}
