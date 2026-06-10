package configurations

import (
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	_ "github.com/6022-labs/agentic-agent-node/src/gateway_host/docs"
	"github.com/6022-labs/agentic-agent-node/src/gateway_host/settings"
	"github.com/6022-labs/agentic-agent-node/src/gateway_mvc/middlewares"
	"github.com/gofiber/contrib/fiberzap/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/swagger"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

const appName = "AgentNode-Gateway"

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
		AppName:   appName,
		Network:   "tcp",
		BodyLimit: hostSettings.BodyLimitBytes,
	})
}

type registerRoutesParams struct {
	dig.In

	App         *fiber.App
	Logger      *zap.Logger
	JwtAuth     *middlewares.JwtAuthMiddleware
	Controllers []mvc.AuthAwareControllerInterface `group:"gateway_controllers"`
}

func registerRoutes(p registerRoutesParams) {
	p.App.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "*",
	}))

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
