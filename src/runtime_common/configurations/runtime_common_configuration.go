package configurations

import (
	"github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	"github.com/6022-labs/agentic-agent-node/src/runtime_common/use_cases"
	"go.uber.org/dig"
)

// AddRuntimeCommonConfiguration registers the node's chain binding: the chain
// id, the well-known contract addresses, and the use case serving them to the
// gateway. The /internal endpoint lives in runtime_common_mvc.
func AddRuntimeCommonConfiguration(container *dig.Container) {
	settingsProviders := []any{
		settings.NewBlockchainSettings,
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
