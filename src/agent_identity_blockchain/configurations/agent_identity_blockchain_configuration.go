package configurations

import (
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/factory"
	blockchain_services "github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/services"
	blockchain_settings "github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/settings"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

func AddAgentIdentityBlockchainConfiguration(container *dig.Container) {
	if err := container.Provide(blockchain_settings.NewRpcSettings); err != nil {
		panic(err)
	}

	if err := container.Provide(newEthClient); err != nil {
		panic(err)
	}

	if err := container.Provide(factory.NewTransactorFactory); err != nil {
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

	if err := container.Provide(
		blockchain_services.NewAgentMinter,
		dig.As(new(identity_interfaces.AgentMinterInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		blockchain_services.NewToken6022Requester,
		dig.As(new(identity_interfaces.Token6022RequesterInterface)),
	); err != nil {
		panic(err)
	}
}

func newEthClient(rpcSettings *blockchain_settings.RpcSettings, logger *zap.Logger) *ethclient.Client {
	client, err := ethclient.Dial(rpcSettings.HttpUrl)
	if err != nil {
		logger.Fatal("failed to dial EVM RPC", zap.Error(err))
	}
	return client
}
