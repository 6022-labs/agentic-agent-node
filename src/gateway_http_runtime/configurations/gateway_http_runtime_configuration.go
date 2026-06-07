package configurations

import (
	"net/http"

	common_interfaces "github.com/6022-labs/agentic-agent-node/src/common/interfaces"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/settings"
	"go.uber.org/dig"
)

// newRuntimeClient builds the generated runtime client (oapi-codegen) bound to
// the configured base URL and timeout.
func newRuntimeClient(s *settings.RuntimeClientSettings) (clients.ClientWithResponsesInterface, error) {
	httpClient := &http.Client{Timeout: s.Timeout}
	return clients.NewClientWithResponses(s.BaseUrl, clients.WithHTTPClient(httpClient))
}

func AddGatewayHttpRuntimeConfiguration(container *dig.Container) {
	if err := container.Provide(settings.NewRuntimeClientSettings); err != nil {
		panic(err)
	}
	if err := container.Provide(newRuntimeClient); err != nil {
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
	if err := container.Provide(
		services.NewHttpAgentRegistrationClient,
		dig.As(new(gateway_interfaces.AgentRegistrationClientInterface)),
	); err != nil {
		panic(err)
	}
}
