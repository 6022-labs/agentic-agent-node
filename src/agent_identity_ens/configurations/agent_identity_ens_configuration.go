package configurations

import (
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_ens/factory"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_ens/services"
	"go.uber.org/dig"
)

// AddAgentIdentityEnsConfiguration registers the on-chain ENS translator
// (AgentEnsRegistryRequester) with this adapter's own RPC clients + transactor.
func AddAgentIdentityEnsConfiguration(container *dig.Container) {
	if err := container.Provide(factory.NewEthClientFactory); err != nil {
		panic(err)
	}
	if err := container.Provide(factory.NewTransactorFactory); err != nil {
		panic(err)
	}

	if err := container.Provide(
		services.NewAgentEnsRegistryRequester,
		dig.As(new(identity_interfaces.EnsRegistrarInterface)),
	); err != nil {
		panic(err)
	}
}
