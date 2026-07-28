package services_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/services"
	"github.com/stretchr/testify/assert"
)

type WhenGettingAgentInformationTestingSuite struct {
	fixture *chainFixture
	sut     *services.AgentCollectionRequester
}

func WhenGettingAgentInformationBeforeEach(t *testing.T) *WhenGettingAgentInformationTestingSuite {
	fixture := newChainFixture(t)
	return &WhenGettingAgentInformationTestingSuite{
		fixture: fixture,
		sut:     fixture.agentCollectionRequester(),
	}
}

func TestWhenGettingAgentInformation(t *testing.T) {
	t.Parallel()

	t.Run("Given a minted agent", func(t *testing.T) {
		t.Parallel()

		agentName := "described-agent"

		t.Run("Should return its name, creator, images and attributes", func(t *testing.T) {
			t.Parallel()
			suite := WhenGettingAgentInformationBeforeEach(t)
			tokenId := suite.fixture.mintAgent(t, agentName)

			info, err := suite.sut.InformationOf(suite.fixture.chainId(), suite.fixture.stack.Collection, tokenId)

			assert.NoError(t, err)
			assert.Equal(t, agentName, info.Name)
			assert.Equal(t, suite.fixture.anvilAccountSettings.Address, info.Creator)
			assert.Nil(t, info.CloneOf)
			assert.Equal(t, "bafybeibsulc7sgpck3w7ur7eqa6fwaazcsswn3dg3rfl65zka3knajx43i", info.Images["default"])
			assert.Equal(t, "facilitator", info.Attributes["role"])
			assert.Equal(t, "integration testing", info.Attributes["mission"])
		})
	})
}
