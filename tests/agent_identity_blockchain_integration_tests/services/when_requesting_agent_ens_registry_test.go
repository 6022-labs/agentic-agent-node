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

type WhenRequestingAgentEnsRegistryTestingSuite struct {
	fixture *chainFixture
	sut     *services.AgentEnsRegistryRequester
}

func WhenRequestingAgentEnsRegistryBeforeEach(t *testing.T) *WhenRequestingAgentEnsRegistryTestingSuite {
	fixture := newChainFixture(t)
	return &WhenRequestingAgentEnsRegistryTestingSuite{
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

func TestWhenRequestingAgentEnsRegistry(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	runtimeUrl := "https://agent.example.com"

	t.Run("Given a minted agent with no published records", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report the subnode as not provisioned", func(t *testing.T) {
			t.Parallel()
			suite := WhenRequestingAgentEnsRegistryBeforeEach(t)
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

		t.Run("Should report it provisioned and read the url back", func(t *testing.T) {
			t.Parallel()
			suite := WhenRequestingAgentEnsRegistryBeforeEach(t)
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

			text, err := suite.sut.Text(
				ctx,
				suite.fixture.chainId(),
				suite.fixture.stack.Collection,
				agentName,
				identity_interfaces.UrlRecordKey,
			)
			assert.NoError(t, err)
			assert.Equal(t, runtimeUrl, text)
		})

		t.Run("Should overwrite an existing record", func(t *testing.T) {
			t.Parallel()
			suite := WhenRequestingAgentEnsRegistryBeforeEach(t)
			tokenId := suite.fixture.mintAgent(t, agentName)

			err := suite.sut.InitRecords(
				ctx,
				suite.fixture.chainId(),
				suite.fixture.stack.Collection,
				new(big.Int).SetUint64(tokenId),
				[]identity_interfaces.EnsRecord{{Key: identity_interfaces.UrlRecordKey, Value: runtimeUrl}},
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
				agentName,
				identity_interfaces.UrlRecordKey,
			)
			assert.NoError(t, err)
			assert.Equal(t, "https://moved.example.com", text)
		})
	})

	t.Run("Given an init batch missing the mandatory url key", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail instead of provisioning a urlless subnode", func(t *testing.T) {
			t.Parallel()
			suite := WhenRequestingAgentEnsRegistryBeforeEach(t)
			tokenId := suite.fixture.mintAgent(t, "urlless-agent")

			err := suite.sut.InitRecords(
				ctx,
				suite.fixture.chainId(),
				suite.fixture.stack.Collection,
				new(big.Int).SetUint64(tokenId),
				[]identity_interfaces.EnsRecord{{Key: "description", Value: "no url here"}},
			)

			assert.Error(t, err)
		})
	})

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

			assert.Error(t, err)
		})
	})
}
