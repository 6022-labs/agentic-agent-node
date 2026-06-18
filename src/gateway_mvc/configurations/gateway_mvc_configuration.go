package configurations

import (
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/6022-labs/agentic-agent-node/src/gateway_mvc/core"
	core_middlewares "github.com/6022-labs/agentic-agent-node/src/gateway_mvc/core/middlewares"
	"github.com/6022-labs/agentic-agent-node/src/gateway_mvc/identity"
	"github.com/6022-labs/agentic-agent-node/src/gateway_mvc/knowledge"
	"github.com/6022-labs/agentic-agent-node/src/gateway_mvc/payment"
	"github.com/6022-labs/agentic-agent-node/src/gateway_mvc/reasoning"
	reasoning_middlewares "github.com/6022-labs/agentic-agent-node/src/gateway_mvc/reasoning/middlewares"
	middleware_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway_mvc/reasoning/middlewares/interfaces"
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
	provideController(container, payment.NewPaymentPolicyController)
	provideController(container, payment.NewPaymentRuleController)
}

func addReasoningConfiguration(container *dig.Container) {
	if err := container.Provide(reasoning_middlewares.NewResponseAccessGateMiddleware); err != nil {
		panic(err)
	}

	// One x402 instance exposed under two groups: the access chain and the
	// free-access evaluators the JWT link consults.
	if err := container.Provide(reasoning_middlewares.NewX402AccessMethod); err != nil {
		panic(err)
	}
	if err := container.Provide(
		func(x *reasoning_middlewares.X402AccessMethod) middleware_interfaces.AccessMethodInterface {
			return x
		},
		dig.Group("access_methods"),
	); err != nil {
		panic(err)
	}
	if err := container.Provide(
		func(x *reasoning_middlewares.X402AccessMethod) middleware_interfaces.FreeRuleMatcherInterface {
			return x
		},
		dig.Group("free_rule_matchers"),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(reasoning_middlewares.NewJwtFreeAccessMethod); err != nil {
		panic(err)
	}
	if err := container.Provide(
		func(x *reasoning_middlewares.JwtFreeAccessMethod) middleware_interfaces.AccessMethodInterface {
			return x
		},
		dig.Group("access_methods"),
	); err != nil {
		panic(err)
	}

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
