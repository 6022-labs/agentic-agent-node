package services_test

import (
	"testing"

	agent_enums "github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/services"
	"github.com/stretchr/testify/assert"
)

type WhenResolvingTokenIdFromAddressTestingSuite struct {
	fixture *chainFixture
	sut     *services.AgentCollectionRequester
}

func WhenResolvingTokenIdFromAddressBeforeEach(t *testing.T) *WhenResolvingTokenIdFromAddressTestingSuite {
	fixture := newChainFixture(t)
	return &WhenResolvingTokenIdFromAddressTestingSuite{
		fixture: fixture,
		sut:     fixture.agentCollectionRequester(),
	}
}

func TestWhenResolvingTokenIdFromAddress(t *testing.T) {
	t.Parallel()

	t.Run("Given a minted agent", func(t *testing.T) {
		t.Parallel()

		t.Run("Should resolve its token id from the registered evm address", func(t *testing.T) {
			t.Parallel()
			suite := WhenResolvingTokenIdFromAddressBeforeEach(t)
			tokenId := suite.fixture.mintAgent(t, "addressed-agent")

			resolved, err := suite.sut.AddressToTokenId(
				suite.fixture.chainId(),
				suite.fixture.stack.Collection,
				agent_enums.AddressTypeEvm,
				suite.fixture.anvilAccountSettings.Address.Hex(),
			)

			assert.NoError(t, err)
			assert.Equal(t, tokenId, resolved)
		})
	})
}
