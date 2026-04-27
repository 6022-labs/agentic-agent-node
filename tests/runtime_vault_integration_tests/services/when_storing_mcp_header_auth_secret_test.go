package services_test

import (
	"context"
	"testing"

	vault_services "github.com/6022-labs/agentic-agent-node/src/runtime_vault/services"
	vault_settings "github.com/6022-labs/agentic-agent-node/src/runtime_vault/settings"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/container_builders"
	vault "github.com/hashicorp/vault/api"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type WhenStoringMCPHeaderAuthSecretTestingSuite struct {
	mountPath string
	sut       *vault_services.MCPHeaderAuthSecretStorage
}

func WhenStoringMCPHeaderAuthSecretBeforeEach() *WhenStoringMCPHeaderAuthSecretTestingSuite {
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

	return &WhenStoringMCPHeaderAuthSecretTestingSuite{
		mountPath: cfg.KvV1MountPath,
		sut:       vault_services.NewMCPHeaderAuthSecretStorage(storage),
	}
}

func WhenStoringMCPHeaderAuthSecretAfterEach(suite *WhenStoringMCPHeaderAuthSecretTestingSuite) {
	container_builders.GetVaultContainerBuilder().DropMount(suite.mountPath)
}

func TestWhenStoringMCPHeaderAuthSecret(t *testing.T) {
	t.Parallel()

	t.Run("Given no header value has been saved", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report not-found from GetValue", func(t *testing.T) {
			t.Parallel()
			suite := WhenStoringMCPHeaderAuthSecretBeforeEach()
			defer WhenStoringMCPHeaderAuthSecretAfterEach(suite)

			value, found, err := suite.sut.GetValue(context.Background(), uuid.New(), "Authorization")
			assert.NoError(t, err)
			assert.False(t, found)
			assert.Empty(t, value)
		})
	})

	t.Run("Given a header value has been saved", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return it from GetValue", func(t *testing.T) {
			t.Parallel()
			suite := WhenStoringMCPHeaderAuthSecretBeforeEach()
			defer WhenStoringMCPHeaderAuthSecretAfterEach(suite)

			ctx := context.Background()
			connectionId := uuid.New()
			assert.NoError(t, suite.sut.SaveValue(ctx, connectionId, "Authorization", "Bearer abc"))

			value, found, err := suite.sut.GetValue(ctx, connectionId, "Authorization")
			assert.NoError(t, err)
			assert.True(t, found)
			assert.Equal(t, "Bearer abc", value)
		})
	})

	t.Run("Given two connections share a header name", func(t *testing.T) {
		t.Parallel()

		t.Run("Should keep their values isolated by connectionId", func(t *testing.T) {
			t.Parallel()
			suite := WhenStoringMCPHeaderAuthSecretBeforeEach()
			defer WhenStoringMCPHeaderAuthSecretAfterEach(suite)

			ctx := context.Background()
			a := uuid.New()
			b := uuid.New()
			assert.NoError(t, suite.sut.SaveValue(ctx, a, "Authorization", "Bearer aaa"))
			assert.NoError(t, suite.sut.SaveValue(ctx, b, "Authorization", "Bearer bbb"))

			va, _, err := suite.sut.GetValue(ctx, a, "Authorization")
			assert.NoError(t, err)
			assert.Equal(t, "Bearer aaa", va)

			vb, _, err := suite.sut.GetValue(ctx, b, "Authorization")
			assert.NoError(t, err)
			assert.Equal(t, "Bearer bbb", vb)
		})
	})

	t.Run("Given SaveValue is called with empty inputs", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject empty key", func(t *testing.T) {
			t.Parallel()
			suite := WhenStoringMCPHeaderAuthSecretBeforeEach()
			defer WhenStoringMCPHeaderAuthSecretAfterEach(suite)

			err := suite.sut.SaveValue(context.Background(), uuid.New(), "", "value")
			assert.Error(t, err)
		})

		t.Run("Should reject empty value", func(t *testing.T) {
			t.Parallel()
			suite := WhenStoringMCPHeaderAuthSecretBeforeEach()
			defer WhenStoringMCPHeaderAuthSecretAfterEach(suite)

			err := suite.sut.SaveValue(context.Background(), uuid.New(), "Authorization", "")
			assert.Error(t, err)
		})
	})

	t.Run("Given a header value has been saved", func(t *testing.T) {
		t.Parallel()

		t.Run("Should clear it on DeleteValue", func(t *testing.T) {
			t.Parallel()
			suite := WhenStoringMCPHeaderAuthSecretBeforeEach()
			defer WhenStoringMCPHeaderAuthSecretAfterEach(suite)

			ctx := context.Background()
			connectionId := uuid.New()
			assert.NoError(t, suite.sut.SaveValue(ctx, connectionId, "Authorization", "Bearer bye"))

			assert.NoError(t, suite.sut.DeleteValue(ctx, connectionId, "Authorization"))

			_, found, err := suite.sut.GetValue(ctx, connectionId, "Authorization")
			assert.NoError(t, err)
			assert.False(t, found)
		})
	})
}
