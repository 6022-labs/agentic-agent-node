package configurations

import (
	"github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	"github.com/6022-labs/agentic-agent-node/src/runtime_common/use_cases"
	"go.uber.org/dig"
)

// AddRuntimeCommonConfiguration registers the reachable chains, the per-chain
// contract registries, and the use case serving them to the gateway. Each
// blockchain adapter owns its own RPC clients.
func AddRuntimeCommonConfiguration(container *dig.Container) {
	settingsProviders := []any{
		settings.NewChainsSettings,
		settings.NewAgentCollectionsManagersSettings,
		settings.NewAgentCollectionCreatorsSettings,
		settings.NewToken6022Settings,
		settings.NewAgentEnsRegistriesSettings,
		settings.NewPaymentFacilitatorsSettings,
	}
	for _, p := range settingsProviders {
		if err := container.Provide(p); err != nil {
			panic(err)
		}
	}

	if err := container.Provide(use_cases.NewGetBlockchainConfig); err != nil {
		panic(err)
	}
}
