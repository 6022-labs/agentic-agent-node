package configurations

import (
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/factory"
	blockchain_services "github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/services"
	"go.uber.org/dig"
)

func AddAgentIdentityBlockchainConfiguration(container *dig.Container) {
	if err := container.Provide(factory.NewEthClientFactory); err != nil {
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

	if err := container.Provide(
		blockchain_services.NewAgentEnsRegistryRequester,
		dig.As(new(identity_interfaces.EnsRegistrarInterface)),
	); err != nil {
		panic(err)
	}
}
