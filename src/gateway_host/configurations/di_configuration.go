package configurations

import (
	agent_runtime_configurations "github.com/6022-labs/agentic-agent-node/src/agent_runtime/configurations"
	agent_runtime_chain_configurations "github.com/6022-labs/agentic-agent-node/src/agent_runtime_chain/configurations"
	agent_runtime_vault_configurations "github.com/6022-labs/agentic-agent-node/src/agent_runtime_vault/configurations"
	gateway_configurations "github.com/6022-labs/agentic-agent-node/src/gateway/configurations"
	"github.com/6022-labs/agentic-agent-node/src/gateway_host/settings"
	gateway_mvc_configurations "github.com/6022-labs/agentic-agent-node/src/gateway_mvc/configurations"

	"github.com/knadh/koanf/v2"
	"go.uber.org/dig"
)

func ConfigureDI(config *koanf.Koanf) *dig.Container {
	container := dig.New()

	if err := container.Provide(settings.NewHostSettings); err != nil {
		panic(err)
	}
	if err := container.Provide(func() *koanf.Koanf {
		return config
	}); err != nil {
		panic(err)
	}

	agent_runtime_vault_configurations.AddAgentRuntimeVaultConfiguration(container)
	agent_runtime_chain_configurations.AddAgentRuntimeChainConfiguration(container)
	agent_runtime_configurations.AddAgentRuntimeConfiguration(container)
	gateway_configurations.AddGatewayConfiguration(container)
	gateway_mvc_configurations.AddGatewayMvcConfiguration(container)

	return container
}
