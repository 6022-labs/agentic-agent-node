package configurations

import (
	"github.com/6022-labs/agentic-agent-node/src/runtime_mvc"
	"go.uber.org/dig"
)

func AddRuntimeMvcConfiguration(container *dig.Container) {
	provideController(container, runtime_mvc.NewBootstrapController)
	provideController(container, runtime_mvc.NewIdentityController)
	provideController(container, runtime_mvc.NewAuthorizeController)
	provideController(container, runtime_mvc.NewConversationMessagesController)
}

func provideController(container *dig.Container, ctor any) {
	if err := container.Provide(
		ctor,
		dig.Group("runtime_controllers"),
		dig.As(new(runtime_mvc.ControllerInterface)),
	); err != nil {
		panic(err)
	}
}
