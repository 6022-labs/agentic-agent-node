package configurations

import (
	common_interfaces "github.com/6022-labs/agentic-agent-node/src/common/interfaces"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/settings"
	"go.uber.org/dig"
)

func AddGatewayHttpRuntimeConfiguration(container *dig.Container) {
	if err := container.Provide(settings.NewRuntimeClientSettings); err != nil {
		panic(err)
	}
	if err := container.Provide(clients.NewRuntimeHttpClient); err != nil {
		panic(err)
	}
	if err := container.Provide(
		services.NewHttpBootstrapStateReader,
		dig.As(new(common_interfaces.BootstrapStateReaderInterface)),
	); err != nil {
		panic(err)
	}
	if err := container.Provide(
		services.NewHttpAgentIdentityResolver,
		dig.As(new(common_interfaces.AgentIdentityResolverInterface)),
	); err != nil {
		panic(err)
	}
	if err := container.Provide(
		services.NewHttpAuthorizer,
		dig.As(new(common_interfaces.AuthorizerInterface)),
	); err != nil {
		panic(err)
	}
	if err := container.Provide(
		services.NewHttpLLMConfigClient,
		dig.As(new(gateway_interfaces.LLMConfigClientInterface)),
	); err != nil {
		panic(err)
	}
	if err := container.Provide(
		services.NewHttpMCPConfigClient,
		dig.As(new(gateway_interfaces.MCPConfigClientInterface)),
	); err != nil {
		panic(err)
	}
	if err := container.Provide(
		services.NewHttpResponseClient,
		dig.As(new(gateway_interfaces.ResponseClientInterface)),
	); err != nil {
		panic(err)
	}
}
