package configurations

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge_mvc"
	mvc_interfaces "github.com/6022-labs/agentic-agent-node/src/common/mvc/interfaces"
	"go.uber.org/dig"
)

func AddAgentKnowledgeMvcConfiguration(container *dig.Container) {
	provideController(container, agent_knowledge_mvc.NewConversationMessagesController)
	provideController(container, agent_knowledge_mvc.NewConversationsController)
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
