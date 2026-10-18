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

type WhenInitialisingEnsRecordsTestingSuite struct {
	fixture *chainFixture
	sut     *services.AgentEnsRegistryRequester
}

func WhenInitialisingEnsRecordsBeforeEach(t *testing.T) *WhenInitialisingEnsRecordsTestingSuite {
	fixture := newChainFixture(t)
	return &WhenInitialisingEnsRecordsTestingSuite{
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

func TestWhenInitialisingEnsRecords(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given an init batch missing the mandatory url key", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail instead of provisioning a urlless subnode", func(t *testing.T) {
			t.Parallel()
			suite := WhenInitialisingEnsRecordsBeforeEach(t)
			tokenId := suite.fixture.mintAgent(t, "urlless-agent")

			err := suite.sut.InitRecords(
				ctx,
				suite.fixture.chainId(),
				suite.fixture.stack.Collection,
				new(big.Int).SetUint64(tokenId),
				[]identity_interfaces.EnsRecord{{Key: "description", Value: "no url here"}},
			)

			assert.ErrorContains(t, err, "custom error 0xc32be248")
		})
	})

	t.Run("Given an init batch carrying the url key", func(t *testing.T) {
		t.Parallel()

		t.Run("Should publish the url record", func(t *testing.T) {
			t.Parallel()
			suite := WhenInitialisingEnsRecordsBeforeEach(t)
			runtimeUrl := "https://agent.example.com"
			tokenId := suite.fixture.mintAgent(t, "published-agent")

			err := suite.sut.InitRecords(
				ctx,
				suite.fixture.chainId(),
				suite.fixture.stack.Collection,
				new(big.Int).SetUint64(tokenId),
				[]identity_interfaces.EnsRecord{{Key: identity_interfaces.UrlRecordKey, Value: runtimeUrl}},
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
			assert.Equal(t, runtimeUrl, text)
		})
	})
}
