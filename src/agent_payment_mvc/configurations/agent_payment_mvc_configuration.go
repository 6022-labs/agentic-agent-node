package configurations

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_payment_mvc"
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"go.uber.org/dig"
)

func AddAgentPaymentMvcConfiguration(container *dig.Container) {
	provideController(container, agent_payment_mvc.NewPaymentPolicyController)
	provideController(container, agent_payment_mvc.NewPaymentRuleController)
	provideController(container, agent_payment_mvc.NewPaymentFacilitatorController)
}

func provideController(container *dig.Container, ctor any) {
	if err := container.Provide(
		ctor,
		dig.Group("runtime_controllers"),
		dig.As(new(mvc.ControllerInterface)),
	); err != nil {
		panic(err)
	}
}
