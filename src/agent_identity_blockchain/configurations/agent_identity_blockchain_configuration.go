package configurations

import (
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	blockchain_services "github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/services"
	blockchain_settings "github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/settings"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

func AddAgentIdentityBlockchainConfiguration(container *dig.Container) {
	if err := container.Provide(blockchain_settings.NewBlockchainSettings); err != nil {
		panic(err)
	}

	if err := container.Provide(blockchain_settings.NewAgentCollectionsManagersSettings); err != nil {
		panic(err)
	}

	if err := container.Provide(newEthClient); err != nil {
		panic(err)
	}

	if err := container.Provide(
		blockchain_services.NewAgentCollectionRequester,
		dig.As(new(identity_interfaces.AgentCollectionRequesterInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		blockchain_services.NewAgentCollectionsManagerRequester,
		dig.As(new(identity_interfaces.AgentCollectionsManagerRequesterInterface)),
	); err != nil {
		panic(err)
	}
}

func newEthClient(settings *blockchain_settings.BlockchainSettings, logger *zap.Logger) *ethclient.Client {
	client, err := ethclient.Dial(settings.RpcHttpUrl)
	if err != nil {
		logger.Fatal("failed to dial EVM RPC", zap.Error(err))
	}
	return client
}
