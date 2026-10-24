package configurations

import (
	mvc_interfaces "github.com/6022-labs/agentic-agent-node/src/common/mvc/interfaces"
	mvc_middlewares "github.com/6022-labs/agentic-agent-node/src/common/mvc/middlewares"
	"github.com/6022-labs/agentic-agent-node/src/common/settings"
	_ "github.com/6022-labs/agentic-agent-node/src/gateway_host/docs"
	"github.com/6022-labs/agentic-agent-node/src/gateway_mvc/core/middlewares"
	fiberotel "github.com/gofiber/contrib/v3/otel"
	"github.com/gofiber/contrib/v3/swaggo"
	fiberzap "github.com/gofiber/contrib/v3/zap"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

const AppName = "6022-AgentNode-Gateway"

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
	AgentTelemetry    *middlewares.AgentTelemetryContextMiddleware
	JwtAuth           *middlewares.JwtAuthMiddleware
	RequestMetrics    *mvc_middlewares.ApiRequestMetricsMiddleware
	RequestFailureLog *mvc_middlewares.RequestFailureLogMiddleware
	Controllers       []mvc_interfaces.AuthAwareControllerInterface `group:"gateway_controllers"`
}

func registerRoutes(p registerRoutesParams) {
	p.App.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{"*"},
	}))

	p.App.Use(p.AgentTelemetry.Handle)
	// Spans only: ApiRequestMetrics owns http.server.*.
	p.App.Use(fiberotel.Middleware(fiberotel.WithoutMetrics(true)))
	p.App.Use(p.RequestMetrics.Handle)

	p.App.Use(fiberzap.New(fiberzap.Config{
		Logger: p.Logger,
	}))

	p.App.Use(p.RequestFailureLog.Handle)

	p.App.Get("/swagger/*", swaggo.New())

	for _, controller := range p.Controllers {
		controller.RegisterUnauthenticatedRoutes(p.App)
	}

	authenticated := p.App.Group("", p.JwtAuth.Handle)
	for _, controller := range p.Controllers {
		controller.RegisterAuthenticatedRoutes(authenticated)
	}
}
