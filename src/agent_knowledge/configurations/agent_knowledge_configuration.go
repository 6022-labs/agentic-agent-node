package configurations

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/metrics"
	metrics_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_knowledge/metrics/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/services"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases"
	"go.uber.org/dig"
)

func AddAgentKnowledgeConfiguration(container *dig.Container) {
	if err := container.Provide(
		metrics.NewKnowledgeMetrics,
		dig.As(new(metrics_interfaces.KnowledgeMetricsInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		services.NewMessageAppender,
		dig.As(new(interfaces.MessageAppenderInterface)),
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
		services.NewConversationLabeler,
		dig.As(new(interfaces.ConversationLabelerInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		services.NewConversationResolver,
		dig.As(new(interfaces.ConversationResolverInterface)),
	); err != nil {
		panic(err)
	}

	useCaseProviders := []any{
		use_cases.NewListConversationMessages,
		use_cases.NewListConversations,
		use_cases.NewAppendConversationMessages,
		use_cases.NewCreateConversation,
		use_cases.NewUpdateConversation,
	}
	for _, p := range useCaseProviders {
		if err := container.Provide(p); err != nil {
			panic(err)
		}
	}
}
