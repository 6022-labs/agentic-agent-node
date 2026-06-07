package configurations

import (
	"github.com/6022-labs/agentic-agent-node/src/gateway_mvc"
	"github.com/6022-labs/agentic-agent-node/src/gateway_mvc/middlewares"
	"go.uber.org/dig"
)

func AddGatewayMvcConfiguration(container *dig.Container) {
	if err := container.Provide(middlewares.NewJwtAuthMiddleware); err != nil {
		panic(err)
	}

	provideController(container, gateway_mvc.NewHealthController)
	provideController(container, gateway_mvc.NewAuthController)
	provideController(container, gateway_mvc.NewResponseController)
	provideController(container, gateway_mvc.NewLlmConfigController)
	provideController(container, gateway_mvc.NewMcpConfigController)
	provideController(container, gateway_mvc.NewAgentRegistrationController)
	provideController(container, gateway_mvc.NewConversationsController)
}

func provideController(container *dig.Container, ctor any) {
	if err := container.Provide(
		ctor,
		dig.Group("controllers"),
		dig.As(new(gateway_mvc.ControllerInterface)),
	); err != nil {
		panic(err)
	}
}
