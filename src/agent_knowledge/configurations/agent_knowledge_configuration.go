package configurations

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases"
	"go.uber.org/dig"
)

func AddAgentKnowledgeConfiguration(container *dig.Container) {
	if err := container.Provide(use_cases.NewAppendConversationMessages); err != nil {
		panic(err)
	}
	if err := container.Provide(use_cases.NewListConversationMessages); err != nil {
		panic(err)
	}
}
