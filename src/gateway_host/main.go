package main

import (
	"context"
	"fmt"

	agent_runtime_services "github.com/6022-labs/agentic-agent-node/src/agent_runtime/services"
	"github.com/6022-labs/agentic-agent-node/src/gateway_host/configurations"
	"github.com/6022-labs/agentic-agent-node/src/gateway_host/settings"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	_ = godotenv.Load(".env")

	config, err := configurations.LoadKoanfConfig()
	if err != nil {
		panic(err)
	}

	container := configurations.ConfigureDI(config)
	configurations.ConfigureLogging(container)
	configurations.ConfigureServer(container)

	// Kick off the bootstrap runner in a goroutine. Server boots immediately;
	// gated routes return 503 until the runner moves the state to "ready".
	if err := container.Invoke(func(runner *agent_runtime_services.BootstrapRunner, logger *zap.Logger) {
		go func() {
			if err := runner.Run(context.Background()); err != nil {
				logger.Error("bootstrap runner failed", zap.Error(err))
			}
		}()
	}); err != nil {
		panic(err)
	}

	err = container.Invoke(func(app *fiber.App, logger *zap.Logger, hostSettings *settings.HostSettings) error {
		apiPort := fmt.Sprintf("%d", hostSettings.ApiPort)
		logger.Info("Starting server on port " + apiPort)
		return app.Listen(":" + apiPort)
	})

	if err != nil {
		panic(err)
	}
}
