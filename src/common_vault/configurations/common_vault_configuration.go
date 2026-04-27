package configurations

import (
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	vault_services "github.com/6022-labs/agentic-agent-node/src/common_vault/services"
	vault_interfaces "github.com/6022-labs/agentic-agent-node/src/common_vault/services/interfaces"
	vault_settings "github.com/6022-labs/agentic-agent-node/src/common_vault/settings"
	vault "github.com/hashicorp/vault/api"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

func AddCommonVaultConfiguration(container *dig.Container) {
	if err := container.Provide(vault_settings.NewVaultSettings); err != nil {
		panic(err)
	}

	if err := container.Provide(newVaultClient); err != nil {
		panic(err)
	}

	if err := container.Provide(
		vault_services.NewKvV1SecretStorage,
		dig.As(new(vault_interfaces.KvV1SecretStorageInterface)),
	); err != nil {
		panic(err)
	}

	if err := container.Provide(
		vault_services.NewWalletCredentialsStorage,
		dig.As(new(identity_interfaces.WalletCredentialsStorageInterface)),
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
