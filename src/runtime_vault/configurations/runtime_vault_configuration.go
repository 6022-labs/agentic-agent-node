package configurations

import (
	runtime_common_interfaces "github.com/6022-labs/agentic-agent-node/src/runtime_common/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/runtime_vault/metrics"
	metrics_interfaces "github.com/6022-labs/agentic-agent-node/src/runtime_vault/metrics/interfaces"
	vault_services "github.com/6022-labs/agentic-agent-node/src/runtime_vault/services"
	vault_settings "github.com/6022-labs/agentic-agent-node/src/runtime_vault/settings"
	vault "github.com/hashicorp/vault/api"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

func AddRuntimeVaultConfiguration(container *dig.Container) {
	if err := container.Provide(vault_settings.NewVaultSettings); err != nil {
		panic(err)
	}

	if err := container.Provide(
		metrics.NewVaultMetrics,
		dig.As(new(metrics_interfaces.VaultMetricsInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(newVaultClient); err != nil {
		panic(err)
	}

	if err := container.Provide(
		vault_services.NewKvV1SecretStorage,
		dig.As(new(runtime_common_interfaces.SecretStorageInterface)),
	); err != nil {
		panic(err)
	}
}

func newVaultClient(settings *vault_settings.VaultSettings, logger *zap.Logger) *vault.Client {
	cfg := vault.DefaultConfig()
	cfg.Address = settings.Address

	client, err := vault.NewClient(cfg)
	if err != nil {
		logger.Fatal("failed to create vault client", zap.Error(err))
	}
	client.SetToken(settings.Token)
	return client
}
