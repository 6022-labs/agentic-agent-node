package configurations

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_mvc"
	mvc_interfaces "github.com/6022-labs/agentic-agent-node/src/common/mvc/interfaces"
	"go.uber.org/dig"
)

func AddAgentReasoningMvcConfiguration(container *dig.Container) {
	provideController(container, agent_reasoning_mvc.NewLLMConfigController)
	provideController(container, agent_reasoning_mvc.NewMCPConfigController)
	provideController(container, agent_reasoning_mvc.NewResponseController)
	provideController(container, agent_reasoning_mvc.NewAgentGuidanceController)
}

func provideController(container *dig.Container, ctor any) {
	if err := container.Provide(
		ctor,
		dig.Group("runtime_controllers"),
		dig.As(new(mvc_interfaces.ControllerInterface)),
	); err != nil {
		panic(err)
	}
}
