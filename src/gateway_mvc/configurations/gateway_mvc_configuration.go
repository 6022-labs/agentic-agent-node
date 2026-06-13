package configurations

import (
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/6022-labs/agentic-agent-node/src/gateway_mvc"
	"github.com/6022-labs/agentic-agent-node/src/gateway_mvc/middlewares"
	middleware_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway_mvc/middlewares/interfaces"
	"go.uber.org/dig"
)

func AddGatewayMvcConfiguration(container *dig.Container) {
	if err := container.Provide(middlewares.NewJwtAuthMiddleware); err != nil {
		panic(err)
	}
	if err := container.Provide(middlewares.NewPaymentGateMiddleware); err != nil {
		panic(err)
	}
	if err := container.Provide(
		middlewares.NewX402PaymentMethodMiddleware,
		dig.Group("payment_methods"),
		dig.As(new(middleware_interfaces.PaymentMethodInterface)),
	); err != nil {
		panic(err)
	}

	provideController(container, gateway_mvc.NewHealthController)
	provideController(container, gateway_mvc.NewConfigController)
	provideController(container, gateway_mvc.NewAuthController)
	provideController(container, gateway_mvc.NewResponseController)
	provideController(container, gateway_mvc.NewLlmConfigController)
	provideController(container, gateway_mvc.NewMcpConfigController)
	provideController(container, gateway_mvc.NewPaymentPolicyController)
	provideController(container, gateway_mvc.NewPaymentRuleController)
	provideController(container, gateway_mvc.NewAgentRegistrationController)
	provideController(container, gateway_mvc.NewAgentCollectionsController)
	provideController(container, gateway_mvc.NewConversationsController)
}

func provideController(container *dig.Container, ctor any) {
	if err := container.Provide(
		ctor,
		dig.Group("gateway_controllers"),
		dig.As(new(mvc.AuthAwareControllerInterface)),
	); err != nil {
		panic(err)
	}
}
