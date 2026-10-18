package configurations

import (
	identity_middlewares "github.com/6022-labs/agentic-agent-node/src/agent_identity_mvc/middlewares"
	mvc_interfaces "github.com/6022-labs/agentic-agent-node/src/common/mvc/interfaces"
	mvc_middlewares "github.com/6022-labs/agentic-agent-node/src/common/mvc/middlewares"
	"github.com/6022-labs/agentic-agent-node/src/common/settings"
	runtime_common_mvc_middlewares "github.com/6022-labs/agentic-agent-node/src/runtime_common_mvc/middlewares"
	_ "github.com/6022-labs/agentic-agent-node/src/runtime_host/docs"
	fiberotel "github.com/gofiber/contrib/v3/otel"
	"github.com/gofiber/contrib/v3/swaggo"
	fiberzap "github.com/gofiber/contrib/v3/zap"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

const AppName = "6022-AgentNode-Runtime"

func ConfigureServer(container *dig.Container) {
	if err := container.Provide(newHttpServer); err != nil {
		panic(err)
	}
	if err := container.Invoke(registerRoutes); err != nil {
		panic(err)
	}
}

func newHttpServer(hostSettings *settings.HostSettings) *fiber.App {
	return fiber.New(fiber.Config{
		AppName:   AppName,
		BodyLimit: hostSettings.BodyLimitBytes,
	})
}

type registerRoutesParams struct {
	dig.In

	App               *fiber.App
	Logger            *zap.Logger
	AgentTelemetry    *identity_middlewares.AgentTelemetryContextMiddleware
	RequestMetrics    *mvc_middlewares.ApiRequestMetricsMiddleware
	CallerContext     *runtime_common_mvc_middlewares.CallerContextMiddleware
	RequestFailureLog *runtime_common_mvc_middlewares.RequestFailureLogMiddleware
	Controllers       []mvc_interfaces.ControllerInterface `group:"runtime_controllers"`
}

func registerRoutes(p registerRoutesParams) {
	p.App.Use(p.AgentTelemetry.Handle)
	// Spans only: ApiRequestMetrics owns http.server.*.
	p.App.Use(fiberotel.Middleware(fiberotel.WithoutMetrics(true)))
	p.App.Use(p.RequestMetrics.Handle)

	p.App.Use(fiberzap.New(fiberzap.Config{
		Logger: p.Logger,
	}))

	p.App.Use(p.CallerContext.Handle)

	p.App.Use(p.RequestFailureLog.Handle)

	p.App.Get("/swagger/*", swaggo.New())

	for _, controller := range p.Controllers {
		controller.RegisterRoutes(p.App)
	}
}
