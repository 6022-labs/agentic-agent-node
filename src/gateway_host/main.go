package main

import (
	"fmt"

	"github.com/6022-labs/agentic-agent-node/src/common/host_configurations"
	"github.com/6022-labs/agentic-agent-node/src/gateway_host/configurations"
	"github.com/6022-labs/agentic-agent-node/src/gateway_host/settings"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

// @title Agentic Agent Node — Gateway Host API
// @version 1.0
// @description Public HTTP front for one agent: auth (JWT + EIP-191), bootstrap status, config endpoints, /responses. Authenticated routes proxy to runtime_host on the loopback interface.
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	_ = godotenv.Load(".env")

	config, err := host_configurations.LoadKoanfConfig()
	if err != nil {
		panic(err)
	}

	container := configurations.ConfigureDI(config)
	host_configurations.ConfigureLogging(container)
	configurations.ConfigureServer(container)

	err = container.Invoke(func(app *fiber.App, logger *zap.Logger, hostSettings *settings.HostSettings) error {
		apiPort := fmt.Sprintf("%d", hostSettings.ApiPort)
		logger.Info("Starting gateway server on port " + apiPort)
		return app.Listen(":" + apiPort)
	})

	if err != nil {
		panic(err)
	}
}
