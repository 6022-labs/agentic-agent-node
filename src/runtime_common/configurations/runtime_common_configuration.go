package configurations

import (
	"github.com/6022-labs/agentic-agent-node/src/runtime_common/metrics"
	metrics_interfaces "github.com/6022-labs/agentic-agent-node/src/runtime_common/metrics/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/runtime_common/services"
	services_interfaces "github.com/6022-labs/agentic-agent-node/src/runtime_common/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	"github.com/6022-labs/agentic-agent-node/src/runtime_common/use_cases"
	"go.uber.org/dig"
)

// AddRuntimeCommonConfiguration registers the reachable chains, the per-chain
// contract registries, and the use case serving them to the gateway. Each
// blockchain adapter owns its own RPC clients.
func AddRuntimeCommonConfiguration(container *dig.Container) {
	settingsProviders := []any{
		settings.NewEvmChainsSettings,
		settings.NewBtcChainsSettings,
		settings.NewAgentCollectionsManagersSettings,
		settings.NewAgentCollectionCreatorsSettings,
		settings.NewToken6022Settings,
		settings.NewAgentEnsRegistriesSettings,
		settings.NewPaymentReceiveSettings,
	}
	for _, p := range settingsProviders {
		if err := container.Provide(p); err != nil {
			panic(err)
		}
	}

	if err := container.Provide(
		metrics.NewBlockchainRpcMetrics,
		dig.As(new(metrics_interfaces.BlockchainRpcMetricsInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(use_cases.NewGetBlockchainConfig); err != nil {
		panic(err)
	}

	if err := container.Provide(
		services.NewEmojiDecoder,
		dig.As(new(services_interfaces.EmojiDecoderInterface)),
	); err != nil {
		panic(err)
	}
}
