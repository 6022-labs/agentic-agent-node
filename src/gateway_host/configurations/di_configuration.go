package configurations

import (
	gateway_configurations "github.com/6022-labs/agentic-agent-node/src/gateway/configurations"
	gateway_host_settings "github.com/6022-labs/agentic-agent-node/src/gateway_host/settings"
	gateway_http_runtime_configurations "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/configurations"
	gateway_http_x402_configurations "github.com/6022-labs/agentic-agent-node/src/gateway_http_x402/configurations"
	gateway_mvc_configurations "github.com/6022-labs/agentic-agent-node/src/gateway_mvc/configurations"

	"github.com/knadh/koanf/v2"
	"go.uber.org/dig"
)

func ConfigureDI(config *koanf.Koanf) *dig.Container {
	container := dig.New()

	if err := container.Provide(gateway_host_settings.NewHostSettings); err != nil {
		panic(err)
	}
	if err := container.Provide(func() *koanf.Koanf {
		return config
	}); err != nil {
		panic(err)
	}

	gateway_http_runtime_configurations.AddGatewayHttpRuntimeConfiguration(container)
	gateway_http_x402_configurations.AddGatewayHttpX402Configuration(container)
	gateway_configurations.AddGatewayConfiguration(container)
	gateway_mvc_configurations.AddGatewayMvcConfiguration(container)

	return container
}
