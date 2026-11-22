package configurations

import (
	"time"

	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	mvc_interfaces "github.com/6022-labs/agentic-agent-node/src/common/mvc/interfaces"
	mvc_middlewares "github.com/6022-labs/agentic-agent-node/src/common/mvc/middlewares"
	"github.com/6022-labs/agentic-agent-node/src/common/settings"
	core_settings "github.com/6022-labs/agentic-agent-node/src/gateway/core/settings"
	_ "github.com/6022-labs/agentic-agent-node/src/gateway_host/docs"
	"github.com/6022-labs/agentic-agent-node/src/gateway_mvc/core/middlewares"
	reasoning_middlewares "github.com/6022-labs/agentic-agent-node/src/gateway_mvc/reasoning/middlewares"
	fiberotel "github.com/gofiber/contrib/v3/otel"
	"github.com/gofiber/contrib/v3/swaggo"
	fiberzap "github.com/gofiber/contrib/v3/zap"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/limiter"
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

func newHttpServer(
	hostSettings *settings.HostSettings,
	proxySettings *core_settings.ProxySettings,
	errorWriter *mvc.ErrorWriter,
) *fiber.App {
	return fiber.New(fiber.Config{
		AppName:            AppName,
		BodyLimit:          hostSettings.BodyLimitBytes,
		ErrorHandler:       errorWriter.Write,
		ProxyHeader:        proxySettings.Header,
		TrustProxy:         proxySettings.Header != "",
		TrustProxyConfig:   fiber.TrustProxyConfig{Proxies: proxySettings.TrustedProxies},
		EnableIPValidation: true,
	})
}

type registerRoutesParams struct {
	dig.In

	App               *fiber.App
	Logger            *zap.Logger
	RateLimitSettings *core_settings.RateLimitSettings
	CorsSettings      *core_settings.CorsSettings
	AgentTelemetry    *middlewares.AgentTelemetryContextMiddleware
	JwtAuth           *middlewares.JwtAuthMiddleware
	RequestMetrics    *mvc_middlewares.ApiRequestMetricsMiddleware
	Controllers       []mvc_interfaces.AuthAwareControllerInterface `group:"gateway_controllers"`
}

func registerRoutes(p registerRoutesParams) {
	p.App.Use(cors.New(cors.Config{
		AllowOrigins: p.CorsSettings.AllowOrigins,
		AllowHeaders: []string{"*"},
		// x402 clients read the payment headers.
		ExposeHeaders: []string{
			reasoning_middlewares.PaymentRequiredHeader,
			reasoning_middlewares.PaymentResponseHeader,
		},
	}))

	p.App.Use(p.AgentTelemetry.Handle)
	// Spans only: ApiRequestMetrics owns http.server.*.
	p.App.Use(fiberotel.Middleware(fiberotel.WithoutMetrics(true)))
	p.App.Use(p.RequestMetrics.Handle)

	p.App.Use(fiberzap.New(fiberzap.Config{
		Logger: p.Logger,
	}))

	p.App.Use(limiter.New(limiter.Config{
		Max:        p.RateLimitSettings.RequestsPerSecond,
		Expiration: 1 * time.Second,
		LimitReached: func(fiber.Ctx) error {
			return fiber.ErrTooManyRequests
		},
	}))

	p.App.Get("/swagger/*", swaggo.New())

	for _, controller := range p.Controllers {
		controller.RegisterUnauthenticatedRoutes(p.App)
	}

	authenticated := p.App.Group("", p.JwtAuth.Handle)
	for _, controller := range p.Controllers {
		controller.RegisterAuthenticatedRoutes(authenticated)
	}
}
