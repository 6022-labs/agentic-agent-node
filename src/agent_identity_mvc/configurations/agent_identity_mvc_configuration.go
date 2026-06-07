package configurations

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_mvc"
	"github.com/6022-labs/agentic-agent-node/src/runtime_common"
	"go.uber.org/dig"
)

func AddAgentIdentityMvcConfiguration(container *dig.Container) {
	provideController(container, agent_identity_mvc.NewIdentityController)
	provideController(container, agent_identity_mvc.NewAuthorizeController)
	provideController(container, agent_identity_mvc.NewAgentRegistrationController)
}

func provideController(container *dig.Container, ctor any) {
	if err := container.Provide(
		ctor,
		dig.Group("runtime_controllers"),
		dig.As(new(runtime_common.ControllerInterface)),
	); err != nil {
		panic(err)
	}
}
