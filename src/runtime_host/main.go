package main

import (
	"context"
	"fmt"

	agent_identity_services "github.com/6022-labs/agentic-agent-node/src/agent_identity/services"
	"github.com/6022-labs/agentic-agent-node/src/common/host_configurations"
	"github.com/6022-labs/agentic-agent-node/src/runtime_host/configurations"
	"github.com/6022-labs/agentic-agent-node/src/runtime_host/settings"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

// @title Agentic Agent Node — Runtime Host API
// @version 1.0
// @description Internal-only runtime + knowledge surface. Bound to 127.0.0.1; only gateway_host calls it. All endpoints are mounted under /internal.
// @BasePath /
func main() {
	_ = godotenv.Load(".env")

	config, err := host_configurations.LoadKoanfConfig()
	if err != nil {
		panic(err)
	}

	container := configurations.ConfigureDI(config)
	host_configurations.ConfigureLogging(container)
	configurations.ConfigureServer(container)

	// Runtime owns wallet bootstrap. Gateway sees the result via /internal/bootstrap/status.
	if err := container.Invoke(func(runner *agent_identity_services.BootstrapRunner, logger *zap.Logger) {
		go func() {
			if err := runner.Run(context.Background()); err != nil {
				logger.Error("bootstrap runner failed", zap.Error(err))
			}
		}()
	}); err != nil {
		panic(err)
	}

	err = container.Invoke(func(app *fiber.App, logger *zap.Logger, hostSettings *settings.HostSettings) error {
		bind := fmt.Sprintf("%s:%d", hostSettings.ListenAddress, hostSettings.ApiPort)
		logger.Info("Starting runtime server", zap.String("bind", bind))
		return app.Listen(bind)
	})
	if err != nil {
		panic(err)
	}
}
