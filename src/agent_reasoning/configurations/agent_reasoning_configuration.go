package configurations

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/metrics"
	metrics_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/metrics/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/settings"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/traces"
	traces_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/traces/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases"
	"go.uber.org/dig"
)

func AddAgentReasoningConfiguration(container *dig.Container) {
	if err := container.Provide(settings.NewConversationTitlerSettings); err != nil {
		panic(err)
	}

	if err := container.Provide(settings.NewInferenceLoopSettings); err != nil {
		panic(err)
	}

	if err := container.Provide(
		metrics.NewReasoningMetrics,
		dig.As(new(metrics_interfaces.ReasoningMetricsInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		metrics.NewResponseMetrics,
		dig.As(new(metrics_interfaces.ResponseMetricsInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		traces.NewReasoningTracer,
		dig.As(new(traces_interfaces.ReasoningTracerInterface)),
	); err != nil {
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

	if err := container.Provide(
		services.NewSystemPromptBuilder,
		dig.As(new(interfaces.SystemPromptBuilderInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		services.NewConversationTitler,
		dig.As(new(interfaces.ConversationTitlerInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		services.NewResponseDriver,
		dig.As(new(interfaces.ResponseDriverInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		use_cases.NewCreateResponse,
		dig.As(new(use_cases.CreateResponseInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		services.NewLLMSecretStorage,
		dig.As(new(interfaces.LLMSecretStorageInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		services.NewMCPHeaderAuthSecretStorage,
		dig.As(new(interfaces.MCPHeaderAuthSecretStorageInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		services.NewLLMConnectionValidator,
		dig.As(new(interfaces.LLMConnectionValidatorInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(services.NewSelfToolReconciler); err != nil {
		panic(err)
	}

	providers := []any{
		use_cases.NewSetupLLMConnection,
		use_cases.NewGetLLMConnection,
		use_cases.NewDeleteLLMConnection,
		use_cases.NewValidateLLMConnection,
		use_cases.NewDiscoverMCPServerTools,
		use_cases.NewCreateMCPConnection,
		use_cases.NewListMCPConnections,
		use_cases.NewGetMCPConnection,
		use_cases.NewUpdateMCPConnection,
		use_cases.NewDeleteMCPConnection,
		use_cases.NewGetAgentGuidance,
		use_cases.NewSetAgentGuidance,
		use_cases.NewListSelfTools,
		use_cases.NewUpdateSelfTool,
	}
	for _, p := range providers {
		if err := container.Provide(p); err != nil {
			panic(err)
		}
	}
}
