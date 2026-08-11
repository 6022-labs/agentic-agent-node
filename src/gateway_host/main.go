package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/6022-labs/agentic-agent-node/src/common/host_configurations"
	"github.com/6022-labs/agentic-agent-node/src/common/settings"
	"github.com/6022-labs/agentic-agent-node/src/gateway_host/configurations"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

// @title Agentic Agent Node — Gateway Host API
// @version 1.0
// @description Public HTTP front: auth (JWT + EIP-191), config, /responses; proxies to runtime_host on loopback.
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

	shutdownTelemetry := host_configurations.ConfigureTelemetry(container, configurations.AppName)
	defer func() { _ = shutdownTelemetry(context.Background()) }()

	configurations.ConfigureServer(container)

	err = container.Invoke(func(app *fiber.App, logger *zap.Logger, hostSettings *settings.HostSettings) error {
		address := fmt.Sprintf("%s:%d", hostSettings.ListenAddress, hostSettings.ApiPort)
		logger.Info("Starting gateway server", zap.String("address", address))

		serverErrCh := make(chan error, 1)
		go func() {
			serverErrCh <- app.Listen(address)
		}()

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		select {
		case sig := <-sigCh:
			logger.Info("Shutdown signal received", zap.String("signal", sig.String()))

			if err := app.Shutdown(); err != nil {
				logger.Warn("Fiber shutdown returned an error", zap.Error(err))
			}

			if err := <-serverErrCh; err != nil {
				logger.Error("Server stopped with error", zap.Error(err))
			} else {
				logger.Info("Server stopped")
			}
			logger.Info("Shutdown complete")
			_ = logger.Sync()
			return nil

		case err := <-serverErrCh:
			if err != nil {
				return fmt.Errorf("server exited unexpectedly: %w", err)
			}
			return nil
		}
	})

	if err != nil {
		panic(err)
	}
}
