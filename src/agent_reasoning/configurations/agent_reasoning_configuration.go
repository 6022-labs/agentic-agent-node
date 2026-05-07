package configurations

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/settings"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases"
	"go.uber.org/dig"
)

func AddAgentReasoningConfiguration(container *dig.Container) {
	if err := container.Provide(settings.NewInferenceLoopSettings); err != nil {
		panic(err)
	}

	if err := container.Provide(
		services.NewLLMExecutor,
		dig.As(new(interfaces.LLMExecutorInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		services.NewInferenceLoop,
		dig.As(new(interfaces.InferenceLoopInterface)),
	); err != nil {
		panic(err)
	}

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
