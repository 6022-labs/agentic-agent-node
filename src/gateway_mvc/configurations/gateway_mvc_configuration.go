package configurations

import (
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/6022-labs/agentic-agent-node/src/gateway_mvc/core"
	core_middlewares "github.com/6022-labs/agentic-agent-node/src/gateway_mvc/core/middlewares"
	"github.com/6022-labs/agentic-agent-node/src/gateway_mvc/identity"
	"github.com/6022-labs/agentic-agent-node/src/gateway_mvc/knowledge"
	"github.com/6022-labs/agentic-agent-node/src/gateway_mvc/payment"
	payment_middlewares "github.com/6022-labs/agentic-agent-node/src/gateway_mvc/payment/middlewares"
	middleware_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway_mvc/payment/middlewares/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway_mvc/reasoning"
	"go.uber.org/dig"
)

// AddGatewayMvcConfiguration wires the gateway controllers and middlewares by
// delegating to each per-domain sub-configuration.
func AddGatewayMvcConfiguration(container *dig.Container) {
	addCoreConfiguration(container)
	addIdentityConfiguration(container)
	addKnowledgeConfiguration(container)
	addPaymentConfiguration(container)
	addReasoningConfiguration(container)
}

func addCoreConfiguration(container *dig.Container) {
	if err := container.Provide(core_middlewares.NewJwtAuthMiddleware); err != nil {
		panic(err)
	}

	provideController(container, core.NewHealthController)
	provideController(container, core.NewAuthController)
	provideController(container, core.NewConfigController)
}

func addIdentityConfiguration(container *dig.Container) {
	provideController(container, identity.NewAgentRegistrationController)
	provideController(container, identity.NewAgentCollectionsController)
}

func addKnowledgeConfiguration(container *dig.Container) {
	provideController(container, knowledge.NewConversationsController)
}

func addPaymentConfiguration(container *dig.Container) {
	if err := container.Provide(payment_middlewares.NewPaymentGateMiddleware); err != nil {
		panic(err)
	}
	if err := container.Provide(
		payment_middlewares.NewX402PaymentMethodMiddleware,
		dig.Group("payment_methods"),
		dig.As(new(middleware_interfaces.PaymentMethodInterface)),
	); err != nil {
		panic(err)
	}

	provideController(container, payment.NewPaymentPolicyController)
	provideController(container, payment.NewPaymentRuleController)
}

func addReasoningConfiguration(container *dig.Container) {
	provideController(container, reasoning.NewResponseController)
	provideController(container, reasoning.NewLlmConfigController)
	provideController(container, reasoning.NewMcpConfigController)
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
