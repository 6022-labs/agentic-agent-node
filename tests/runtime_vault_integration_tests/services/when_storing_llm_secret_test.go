package services_test

import (
	"context"
	"testing"

	vault_services "github.com/6022-labs/agentic-agent-node/src/runtime_vault/services"
	vault_settings "github.com/6022-labs/agentic-agent-node/src/runtime_vault/settings"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/container_builders"
	vault "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
)

type WhenStoringLLMSecretTestingSuite struct {
	mountPath string
	sut       *vault_services.LLMSecretStorage
}

func WhenStoringLLMSecretBeforeEach() *WhenStoringLLMSecretTestingSuite {
	cfg := container_builders.GetVaultContainerBuilder().Create()

	vaultCfg := vault.DefaultConfig()
	vaultCfg.Address = cfg.Address
	client, err := vault.NewClient(vaultCfg)
	if err != nil {
		panic(err)
	}
	client.SetToken(cfg.Token)

	settings := &vault_settings.VaultSettings{
		Address:       cfg.Address,
		Token:         cfg.Token,
		KvV1MountPath: cfg.KvV1MountPath,
	}
	storage := vault_services.NewKvV1SecretStorage(client, settings)

	return &WhenStoringLLMSecretTestingSuite{
		mountPath: cfg.KvV1MountPath,
		sut:       vault_services.NewLLMSecretStorage(storage),
	}
}

func WhenStoringLLMSecretAfterEach(suite *WhenStoringLLMSecretTestingSuite) {
	container_builders.GetVaultContainerBuilder().DropMount(suite.mountPath)
}

func TestWhenStoringLLMSecret(t *testing.T) {
	t.Parallel()

	t.Run("Given no api key has been saved", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report not-found from Get", func(t *testing.T) {
			t.Parallel()
			suite := WhenStoringLLMSecretBeforeEach()
			defer WhenStoringLLMSecretAfterEach(suite)

			apiKey, found, err := suite.sut.Get(context.Background())
			assert.NoError(t, err)
			assert.False(t, found)
			assert.Empty(t, apiKey)
		})
	})

	t.Run("Given an api key has been saved", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return it from Get", func(t *testing.T) {
			t.Parallel()
			suite := WhenStoringLLMSecretBeforeEach()
			defer WhenStoringLLMSecretAfterEach(suite)

			ctx := context.Background()
			assert.NoError(t, suite.sut.Save(ctx, "sk-secret-123"))

			apiKey, found, err := suite.sut.Get(ctx)
			assert.NoError(t, err)
			assert.True(t, found)
			assert.Equal(t, "sk-secret-123", apiKey)
		})
	})

	t.Run("Given Save is called twice", func(t *testing.T) {
		t.Parallel()

		t.Run("Should overwrite the previous api key", func(t *testing.T) {
			t.Parallel()
			suite := WhenStoringLLMSecretBeforeEach()
			defer WhenStoringLLMSecretAfterEach(suite)

			ctx := context.Background()
			assert.NoError(t, suite.sut.Save(ctx, "sk-old"))
			assert.NoError(t, suite.sut.Save(ctx, "sk-new"))

			apiKey, found, err := suite.sut.Get(ctx)
			assert.NoError(t, err)
			assert.True(t, found)
			assert.Equal(t, "sk-new", apiKey)
		})

		t.Run("Should reject an empty api key", func(t *testing.T) {
			t.Parallel()
			suite := WhenStoringLLMSecretBeforeEach()
			defer WhenStoringLLMSecretAfterEach(suite)

			err := suite.sut.Save(context.Background(), "")
			assert.Error(t, err)
		})
	})

	t.Run("Given an api key has been saved", func(t *testing.T) {
		t.Parallel()

		t.Run("Should clear it on Delete", func(t *testing.T) {
			t.Parallel()
			suite := WhenStoringLLMSecretBeforeEach()
			defer WhenStoringLLMSecretAfterEach(suite)

			ctx := context.Background()
			assert.NoError(t, suite.sut.Save(ctx, "sk-bye"))
			assert.NoError(t, suite.sut.Delete(ctx))

			_, found, err := suite.sut.Get(ctx)
			assert.NoError(t, err)
			assert.False(t, found)
		})
	})
}
