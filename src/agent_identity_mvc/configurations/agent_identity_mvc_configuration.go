package configurations

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_mvc"
	mvc_interfaces "github.com/6022-labs/agentic-agent-node/src/common/mvc/interfaces"
	"go.uber.org/dig"
)

func AddAgentIdentityMvcConfiguration(container *dig.Container) {
	provideController(container, agent_identity_mvc.NewIdentityController)
	provideController(container, agent_identity_mvc.NewAuthorizeController)
	provideController(container, agent_identity_mvc.NewAgentRegistrationController)
	provideController(container, agent_identity_mvc.NewAgentCollectionsController)
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
