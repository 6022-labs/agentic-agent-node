package services_test

import (
	"context"
	"math/big"
	"testing"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/services"
	"github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type WhenCheckingEnsProvisionedTestingSuite struct {
	fixture *chainFixture
	sut     *services.AgentEnsRegistryRequester
}

func WhenCheckingEnsProvisionedBeforeEach(t *testing.T) *WhenCheckingEnsProvisionedTestingSuite {
	fixture := newChainFixture(t)
	return &WhenCheckingEnsProvisionedTestingSuite{
		fixture: fixture,
		sut: services.NewAgentEnsRegistryRequester(
			zap.NewNop(),
			fixture.ethClientFactory,
			fixture.ensRegistriesSettings(),
			fixture.transactorFactory,
			fixture.agentCollectionRequester(),
		),
	}
}

func TestWhenCheckingEnsProvisioned(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given a chain with no configured registry", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail instead of calling a zero address", func(t *testing.T) {
			t.Parallel()
			fixture := newChainFixture(t)
			emptySettings := settings.AgentEnsRegistriesSettings{}
			sut := services.NewAgentEnsRegistryRequester(
				zap.NewNop(),
				fixture.ethClientFactory,
				&emptySettings,
				fixture.transactorFactory,
				fixture.agentCollectionRequester(),
			)

			_, err := sut.Provisioned(ctx, fixture.chainId(), fixture.stack.Collection, "any-agent")

			assert.ErrorContains(t, err, "no AgentEnsRegistry configured for chainId")
		})
	})

	t.Run("Given a minted agent with no published records", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report the subnode as not provisioned", func(t *testing.T) {
			t.Parallel()
			suite := WhenCheckingEnsProvisionedBeforeEach(t)
			suite.fixture.mintAgent(t, "unprovisioned-agent")

			provisioned, err := suite.sut.Provisioned(
				ctx,
				suite.fixture.chainId(),
				suite.fixture.stack.Collection,
				"unprovisioned-agent",
			)

			assert.NoError(t, err)
			assert.False(t, provisioned)
		})
	})

	t.Run("Given records initialised on the agent subnode", func(t *testing.T) {
		t.Parallel()

		agentName := "published-agent"
		runtimeUrl := "https://agent.example.com"

		t.Run("Should report it provisioned", func(t *testing.T) {
			t.Parallel()
			suite := WhenCheckingEnsProvisionedBeforeEach(t)
			tokenId := suite.fixture.mintAgent(t, agentName)

			err := suite.sut.InitRecords(
				ctx,
				suite.fixture.chainId(),
				suite.fixture.stack.Collection,
				new(big.Int).SetUint64(tokenId),
				[]identity_interfaces.EnsRecord{{Key: identity_interfaces.UrlRecordKey, Value: runtimeUrl}},
			)
			assert.NoError(t, err)

			provisioned, err := suite.sut.Provisioned(
				ctx,
				suite.fixture.chainId(),
				suite.fixture.stack.Collection,
				agentName,
			)
			assert.NoError(t, err)
			assert.True(t, provisioned)
		})
	})
}
