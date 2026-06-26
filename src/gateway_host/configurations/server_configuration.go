package configurations

import (
	mvc_interfaces "github.com/6022-labs/agentic-agent-node/src/common/mvc/interfaces"
	mvc_middlewares "github.com/6022-labs/agentic-agent-node/src/common/mvc/middlewares"
	"github.com/6022-labs/agentic-agent-node/src/common/settings"
	_ "github.com/6022-labs/agentic-agent-node/src/gateway_host/docs"
	"github.com/6022-labs/agentic-agent-node/src/gateway_mvc/core/middlewares"
	"github.com/gofiber/contrib/fiberzap/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/swagger"
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
		Network:   "tcp",
		BodyLimit: hostSettings.BodyLimitBytes,
	})
}

type registerRoutesParams struct {
	dig.In

	App            *fiber.App
	Logger         *zap.Logger
	AgentTelemetry *middlewares.AgentTelemetryContextMiddleware
	JwtAuth        *middlewares.JwtAuthMiddleware
	RequestMetrics *mvc_middlewares.ApiRequestMetricsMiddleware
	Controllers    []mvc_interfaces.AuthAwareControllerInterface `group:"gateway_controllers"`
}

func registerRoutes(p registerRoutesParams) {
	p.App.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "*",
	}))

	p.App.Use(p.AgentTelemetry.Handle)
	p.App.Use(p.RequestMetrics.Handle)

	p.App.Use(fiberzap.New(fiberzap.Config{
		Logger: p.Logger,
	}))

	p.App.Get("/swagger/*", swagger.New())

	for _, controller := range p.Controllers {
		controller.RegisterUnauthenticatedRoutes(p.App)
	}

	authenticated := p.App.Group("", p.JwtAuth.Handle)
	for _, controller := range p.Controllers {
		controller.RegisterAuthenticatedRoutes(authenticated)
	}
}
