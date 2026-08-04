package configurations

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_os/metrics"
	metrics_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_os/metrics/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_os/services"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_os/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_os/settings"
	"go.uber.org/dig"
)

func AddAgentReasoningOsConfiguration(container *dig.Container) {
	if err := container.Provide(settings.NewOsSettings); err != nil {
		panic(err)
	}

	if err := container.Provide(settings.NewCommandSettings); err != nil {
		panic(err)
	}

	if err := container.Provide(
		metrics.NewCommandMetrics,
		dig.As(new(metrics_interfaces.CommandMetricsInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		services.NewCommandRunner,
		dig.As(new(interfaces.CommandRunnerInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		services.NewProcessRegistry,
		dig.As(new(interfaces.ProcessRegistryInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(services.NewCommandExecutor); err != nil {
		panic(err)
	}
}
