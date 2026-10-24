package configurations

import (
	common_metrics "github.com/6022-labs/agentic-agent-node/src/common/metrics"
	common_metrics_interfaces "github.com/6022-labs/agentic-agent-node/src/common/metrics/interfaces"
	mvc_interfaces "github.com/6022-labs/agentic-agent-node/src/common/mvc/interfaces"
	mvc_middlewares "github.com/6022-labs/agentic-agent-node/src/common/mvc/middlewares"
	"github.com/6022-labs/agentic-agent-node/src/runtime_common_mvc"
	runtime_common_mvc_middlewares "github.com/6022-labs/agentic-agent-node/src/runtime_common_mvc/middlewares"
	"go.uber.org/dig"
)

func AddRuntimeCommonMvcConfiguration(container *dig.Container) {
	if err := container.Provide(func() common_metrics_interfaces.ApiRequestMetricsInterface {
		return common_metrics.NewApiRequestMetrics("agentic_runtime_api")
	}); err != nil {
		panic(err)
	}
	if err := container.Provide(mvc_middlewares.NewApiRequestMetricsMiddleware); err != nil {
		panic(err)
	}
	if err := container.Provide(runtime_common_mvc_middlewares.NewCallerContextMiddleware); err != nil {
		panic(err)
	}

	if err := container.Provide(mvc_middlewares.NewRequestFailureLogMiddleware); err != nil {
		panic(err)
	}

	provideController(container, runtime_common_mvc.NewBlockchainConfigController)
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
