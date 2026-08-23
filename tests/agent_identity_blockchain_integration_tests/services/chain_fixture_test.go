package services_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	agent_enums "github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/factory"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/services"
	"github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	"github.com/6022-labs/agentic-agent-node/tests/agent_identity_blockchain_integration_tests/fixtures"
	identity_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	identity_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/container_builders"
	runtime_metrics_mocks "github.com/6022-labs/agentic-agent-node/tests/runtime_common_mocks/metrics_mocks/interfaces_mocks"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

// chainFixture is a throwaway chain plus a freshly deployed contract stack; the
// anvil account is the collection's admin, moderator and the node's own wallet,
// so mint and ENS writes all authorize against it.
type chainFixture struct {
	anvilAccountSettings container_builders.AnvilAccountSettings
	client               *ethclient.Client
	stack                *fixtures.AgentContractStack
	ethClientFactory     *factory.EthClientFactory
	transactorFactory    *factory.TransactorFactory
}

func newChainFixture(t *testing.T) *chainFixture {
	anvilAccountSettings := container_builders.GetAnvilContainerBuilder().Create()

	client, err := ethclient.Dial(anvilAccountSettings.HttpUrl)
	if err != nil {
		t.Fatalf("failed to dial anvil: %v", err)
	}
	t.Cleanup(client.Close)

	stack, err := fixtures.DeployAgentStack(context.Background(), client, anvilAccountSettings)
	if err != nil {
		t.Fatalf("failed to deploy contract stack: %v", err)
	}

	mockController := gomock.NewController(t)

	mockBlockchainRpcMetrics := runtime_metrics_mocks.NewMockBlockchainRpcMetricsInterface(mockController)
	mockBlockchainRpcMetrics.EXPECT().
		RecordRequest(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes()

	agent := &models.Agent{Id: uuid.New()}
	mockAgentRepository := identity_repositories_mocks.NewMockAgentRepositoryInterface(mockController)
	mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil).AnyTimes()

	mockWalletCredentialsStorage := identity_interfaces_mocks.NewMockWalletCredentialsStorageInterface(mockController)
	mockWalletCredentialsStorage.EXPECT().
		GetForId(gomock.Any(), agent.Id.String(), agent_enums.AddressTypeEvm).
		Return(&identity_interfaces.WalletCredentials{
			PrivateKeyHex: anvilAccountSettings.PrivateKeyHex,
			Address:       anvilAccountSettings.Address.Hex(),
		}, true, nil).
		AnyTimes()

	return &chainFixture{
		anvilAccountSettings: anvilAccountSettings,
		client:               client,
		stack:                stack,
		ethClientFactory: factory.NewEthClientFactory(
			evmChainsSettings(anvilAccountSettings),
			mockBlockchainRpcMetrics,
		),
		transactorFactory: factory.NewTransactorFactory(mockWalletCredentialsStorage, mockAgentRepository),
	}
}

func (f *chainFixture) chainId() string {
	return f.anvilAccountSettings.ChainId
}

func (f *chainFixture) collectionsManagersSettings() *settings.AgentCollectionsManagersSettings {
	managersSettings := settings.AgentCollectionsManagersSettings{f.chainId(): f.stack.Manager}
	return &managersSettings
}

func (f *chainFixture) ensRegistriesSettings() *settings.AgentEnsRegistriesSettings {
	registriesSettings := settings.AgentEnsRegistriesSettings{f.chainId(): f.stack.EnsRegistry}
	return &registriesSettings
}

func (f *chainFixture) agentCollectionRequester() *services.AgentCollectionRequester {
	return services.NewAgentCollectionRequester(f.ethClientFactory)
}

func (f *chainFixture) agentCollectionsManagerRequester() *services.AgentCollectionsManagerRequester {
	return services.NewAgentCollectionsManagerRequester(f.ethClientFactory, f.collectionsManagersSettings())
}

func (f *chainFixture) agentMinter() *services.AgentMinter {
	return services.NewAgentMinter(zap.NewNop(), f.ethClientFactory, f.transactorFactory)
}

// mintAgent arranges a minted agent through the contract directly; the minter is
// exercised where it is the subject, not while setting up other tests.
func (f *chainFixture) mintAgent(t *testing.T, agentName string) uint64 {
	if err := fixtures.MintAgent(
		context.Background(),
		f.client,
		f.anvilAccountSettings,
		f.stack.Collection,
		agentName,
		f.anvilAccountSettings.Address,
	); err != nil {
		t.Fatalf("failed to mint %s: %v", agentName, err)
	}

	tokenId, err := fixtures.TokenIdOf(
		context.Background(),
		f.client,
		f.anvilAccountSettings,
		f.stack.Collection,
		agentName,
	)
	if err != nil {
		t.Fatalf("failed to resolve token id for %s: %v", agentName, err)
	}
	return tokenId
}

func mintParams(f *chainFixture, agentName string) identity_interfaces.MintParams {
	return identity_interfaces.MintParams{
		Collection: f.stack.Collection,
		To:         f.anvilAccountSettings.Address,
		Name:       agentName,
		Addresses: []identity_interfaces.MintAddress{
			{AddressType: agent_enums.AddressTypeEvm, Value: f.anvilAccountSettings.Address.Hex()},
		},
		Images: []identity_interfaces.MintKeyValue{
			{Key: "default", Value: "bafybeibsulc7sgpck3w7ur7eqa6fwaazcsswn3dg3rfl65zka3knajx43i"},
		},
		Attributes: []identity_interfaces.MintKeyValue{
			{Key: "role", Value: "facilitator"},
			{Key: "mission", Value: "integration testing"},
		},
	}
}

// evmChainsSettings points the eth client factory at the test's anvil chain.
func evmChainsSettings(anvilAccountSettings container_builders.AnvilAccountSettings) *settings.EvmChainsSettings {
	evmChains := settings.EvmChainsSettings{
		anvilAccountSettings.ChainId: {
			ChainId: anvilAccountSettings.ChainId,
			HttpUrl: anvilAccountSettings.HttpUrl,
		},
	}
	return &evmChains
}
