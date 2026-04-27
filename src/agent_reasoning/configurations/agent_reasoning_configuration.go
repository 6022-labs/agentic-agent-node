package configurations

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases"
	"go.uber.org/dig"
)

func AddAgentReasoningConfiguration(container *dig.Container) {
	providers := []any{
		use_cases.NewSetupLLMConnection,
		use_cases.NewGetLLMConnection,
		use_cases.NewDeleteLLMConnection,
		use_cases.NewCreateMCPConnection,
		use_cases.NewListMCPConnections,
		use_cases.NewGetMCPConnection,
		use_cases.NewUpdateMCPConnection,
		use_cases.NewDeleteMCPConnection,
	}
	for _, p := range providers {
		if err := container.Provide(p); err != nil {
			panic(err)
		}
	}
}
