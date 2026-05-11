package configurations

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/services"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases"
	"go.uber.org/dig"
)

func AddAgentKnowledgeConfiguration(container *dig.Container) {
	if err := container.Provide(
		services.NewListConversationMessages,
		dig.As(new(interfaces.ListConversationMessagesInterface)),
	); err != nil {
		panic(err)
	}
	if err := container.Provide(
		services.NewAppendConversationMessages,
		dig.As(new(interfaces.AppendConversationMessagesInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		services.NewConversationHistory,
		dig.As(new(interfaces.ConversationHistoryInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		use_cases.NewListConversationMessages,
		dig.As(new(use_cases.ListConversationMessagesInterface)),
	); err != nil {
		panic(err)
	}
	if err := container.Provide(
		use_cases.NewAppendConversationMessages,
		dig.As(new(use_cases.AppendConversationMessagesInterface)),
	); err != nil {
		panic(err)
	}
}
