package configurations

import (
	identity_middlewares "github.com/6022-labs/agentic-agent-node/src/agent_identity_mvc/middlewares"
	mvc_interfaces "github.com/6022-labs/agentic-agent-node/src/common/mvc/interfaces"
	mvc_middlewares "github.com/6022-labs/agentic-agent-node/src/common/mvc/middlewares"
	"github.com/6022-labs/agentic-agent-node/src/common/settings"
	_ "github.com/6022-labs/agentic-agent-node/src/runtime_host/docs"
	"github.com/gofiber/contrib/fiberzap/v2"
	"github.com/gofiber/contrib/otelfiber/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
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
		Network:   "tcp",
		BodyLimit: hostSettings.BodyLimitBytes,
	})
}

type registerRoutesParams struct {
	dig.In

	App            *fiber.App
	Logger         *zap.Logger
	AgentTelemetry *identity_middlewares.AgentTelemetryContextMiddleware
	RequestMetrics *mvc_middlewares.ApiRequestMetricsMiddleware
	Controllers    []mvc_interfaces.ControllerInterface `group:"runtime_controllers"`
}

func registerRoutes(p registerRoutesParams) {
	p.App.Use(p.AgentTelemetry.Handle)
	// Spans only: ApiRequestMetrics owns http.server.*.
	p.App.Use(otelfiber.Middleware(otelfiber.WithoutMetrics(true)))
	p.App.Use(p.RequestMetrics.Handle)

	p.App.Use(fiberzap.New(fiberzap.Config{
		Logger: p.Logger,
	}))

	p.App.Use(func(ctx *fiber.Ctx) error {
		err := ctx.Next()
		status := ctx.Response().StatusCode()

		if status < fiber.StatusBadRequest || status == fiber.StatusNotFound {
			return err
		}
		fields := []zap.Field{
			zap.String("method", ctx.Method()),
			zap.String("path", ctx.OriginalURL()),
			zap.Int("status", status),
			zap.ByteString("response", ctx.Response().Body()),
		}
		if err != nil {
			fields = append(fields, zap.Error(err))
		}
		if status >= fiber.StatusInternalServerError {
			p.Logger.Error("runtime request failed", fields...)
		} else {
			p.Logger.Warn("runtime request rejected", fields...)
		}
		return err
	})

	p.App.Get("/swagger/*", swagger.New())

	for _, controller := range p.Controllers {
		controller.RegisterRoutes(p.App)
	}
}
