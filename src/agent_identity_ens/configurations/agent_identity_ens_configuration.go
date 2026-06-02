package configurations

import (
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_ens/services"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_ens/settings"
	"go.uber.org/dig"
)

// AddAgentIdentityEnsConfiguration registers the boot-time ENS writer for the
// agent's runtime `url`. AgentEnsRegistryRequester is
// invoked by BootstrapRunner via the EnsRuntimeRecordsPublisherInterface
// after the wallet is ready.
func AddAgentIdentityEnsConfiguration(container *dig.Container) {
	if err := container.Provide(settings.NewAgentIdentityEnsSettings); err != nil {
		panic(err)
	}

	if err := container.Provide(settings.NewAgentEnsRegistriesSettings); err != nil {
		panic(err)
	}

	if err := container.Provide(
		services.NewAgentEnsRegistryRequester,
		dig.As(new(identity_interfaces.EnsRuntimeRecordsPublisherInterface)),
	); err != nil {
		panic(err)
	}
}
