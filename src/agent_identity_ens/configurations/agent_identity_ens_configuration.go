package configurations

import (
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_ens/services"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_ens/settings"
	"go.uber.org/dig"
)

// AddAgentIdentityEnsConfiguration registers the on-chain ENS translator
// (AgentEnsRegistryRequester); the runtime URL value it publishes is owned by
// the agent_identity domain.
func AddAgentIdentityEnsConfiguration(container *dig.Container) {
	if err := container.Provide(settings.NewEnsBlockchainSettings); err != nil {
		panic(err)
	}

	if err := container.Provide(settings.NewAgentEnsRegistriesSettings); err != nil {
		panic(err)
	}

	if err := container.Provide(
		services.NewAgentEnsRegistryRequester,
		dig.As(new(identity_interfaces.EnsRegistrarInterface)),
	); err != nil {
		panic(err)
	}
}
