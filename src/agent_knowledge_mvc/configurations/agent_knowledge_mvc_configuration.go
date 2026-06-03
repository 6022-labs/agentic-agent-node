package configurations

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge_mvc"
	"github.com/6022-labs/agentic-agent-node/src/runtime_common"
	"go.uber.org/dig"
)

func AddAgentKnowledgeMvcConfiguration(container *dig.Container) {
	provideController(container, agent_knowledge_mvc.NewConversationMessagesController)
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
