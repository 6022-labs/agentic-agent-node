package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

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

	bootstrapCtx, cancelBootstrap := context.WithCancel(context.Background())
	defer cancelBootstrap()

	// Runtime owns wallet bootstrap. Gateway sees the result via /internal/bootstrap/status.
	// BootstrapRunner also publishes ENS runtime records after wallet readiness;
	// ENS failure is non-fatal and handled inside the runner.
	if err := container.Invoke(func(
		bootstrapRunner *agent_identity_services.BootstrapRunner,
		logger *zap.Logger,
	) {
		go func() {
			if err := bootstrapRunner.Run(bootstrapCtx); err != nil {
				logger.Error("bootstrap runner failed", zap.Error(err))
			}
		}()
	}); err != nil {
		panic(err)
	}

	err = container.Invoke(func(app *fiber.App, logger *zap.Logger, hostSettings *settings.HostSettings) error {
		port := fmt.Sprintf(":%d", hostSettings.ApiPort)
		logger.Info("Starting runtime server", zap.String("port", port))

		serverErrCh := make(chan error, 1)
		go func() {
			serverErrCh <- app.Listen(port)
		}()

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		select {
		case sig := <-sigCh:
			logger.Info("Shutdown signal received", zap.String("signal", sig.String()))

			cancelBootstrap()

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
