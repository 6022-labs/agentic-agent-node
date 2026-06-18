package configurations

import (
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/6022-labs/agentic-agent-node/src/runtime_common_mvc"
	"go.uber.org/dig"
)

func AddRuntimeCommonMvcConfiguration(container *dig.Container) {
	provideController(container, runtime_common_mvc.NewBlockchainConfigController)
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
