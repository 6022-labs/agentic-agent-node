package services_test

import (
	"context"
	"math/big"
	"testing"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/services"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type WhenSettingEnsRecordTestingSuite struct {
	fixture *chainFixture
	sut     *services.AgentEnsRegistryRequester
}

func WhenSettingEnsRecordBeforeEach(t *testing.T) *WhenSettingEnsRecordTestingSuite {
	fixture := newChainFixture(t)
	return &WhenSettingEnsRecordTestingSuite{
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

func TestWhenSettingEnsRecord(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given a url record exists", func(t *testing.T) {
		t.Parallel()

		t.Run("Should overwrite it", func(t *testing.T) {
			t.Parallel()
			suite := WhenSettingEnsRecordBeforeEach(t)
			tokenId := suite.fixture.mintAgent(t, "published-agent")

			err := suite.sut.InitRecords(
				ctx,
				suite.fixture.chainId(),
				suite.fixture.stack.Collection,
				new(big.Int).SetUint64(tokenId),
				[]identity_interfaces.EnsRecord{
					{Key: identity_interfaces.UrlRecordKey, Value: "https://agent.example.com"},
				},
			)
			assert.NoError(t, err)

			err = suite.sut.SetRecord(
				ctx,
				suite.fixture.chainId(),
				suite.fixture.stack.Collection,
				new(big.Int).SetUint64(tokenId),
				identity_interfaces.UrlRecordKey,
				"https://moved.example.com",
			)
			assert.NoError(t, err)

			text, err := suite.sut.Text(
				ctx,
				suite.fixture.chainId(),
				suite.fixture.stack.Collection,
				"published-agent",
				identity_interfaces.UrlRecordKey,
			)
			assert.NoError(t, err)
			assert.Equal(t, "https://moved.example.com", text)
		})
	})
}
