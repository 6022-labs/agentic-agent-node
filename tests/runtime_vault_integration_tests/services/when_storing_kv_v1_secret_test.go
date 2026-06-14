package services_test

import (
	"context"
	"errors"
	"testing"

	runtime_common_interfaces "github.com/6022-labs/agentic-agent-node/src/runtime_common/services/interfaces"
	vault_services "github.com/6022-labs/agentic-agent-node/src/runtime_vault/services"
	vault_settings "github.com/6022-labs/agentic-agent-node/src/runtime_vault/settings"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/container_builders"
	vault "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
)

type WhenStoringKvV1SecretTestingSuite struct {
	mountPath string
	sut       *vault_services.KvV1SecretStorage
}

func WhenStoringKvV1SecretBeforeEach() *WhenStoringKvV1SecretTestingSuite {
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

	return &WhenStoringKvV1SecretTestingSuite{
		mountPath: cfg.KvV1MountPath,
		sut:       vault_services.NewKvV1SecretStorage(client, settings),
	}
}

func WhenStoringKvV1SecretAfterEach(suite *WhenStoringKvV1SecretTestingSuite) {
	container_builders.GetVaultContainerBuilder().DropMount(suite.mountPath)
}

func TestWhenStoringKvV1Secret(t *testing.T) {
	t.Parallel()

	t.Run("Given a key that has not been written", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return ErrSecretNotFound on Read", func(t *testing.T) {
			t.Parallel()
			suite := WhenStoringKvV1SecretBeforeEach()
			defer WhenStoringKvV1SecretAfterEach(suite)

			data, err := suite.sut.Read(context.Background(), "missing")
			assert.Nil(t, data)
			assert.True(t, errors.Is(err, runtime_common_interfaces.ErrSecretNotFound))
		})
	})

	t.Run("Given a key that has been written", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return the stored values on Read", func(t *testing.T) {
			t.Parallel()
			suite := WhenStoringKvV1SecretBeforeEach()
			defer WhenStoringKvV1SecretAfterEach(suite)

			ctx := context.Background()
			err := suite.sut.Write(ctx, "alpha", map[string]string{
				"first":  "one",
				"second": "two",
			})
			assert.NoError(t, err)

			data, err := suite.sut.Read(ctx, "alpha")
			assert.NoError(t, err)
			assert.Equal(t, "one", data["first"])
			assert.Equal(t, "two", data["second"])
		})
	})

	t.Run("Given a key that has been overwritten", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return the latest values on Read", func(t *testing.T) {
			t.Parallel()
			suite := WhenStoringKvV1SecretBeforeEach()
			defer WhenStoringKvV1SecretAfterEach(suite)

			ctx := context.Background()
			assert.NoError(t, suite.sut.Write(ctx, "beta", map[string]string{"k": "v1"}))
			assert.NoError(t, suite.sut.Write(ctx, "beta", map[string]string{"k": "v2"}))

			data, err := suite.sut.Read(ctx, "beta")
			assert.NoError(t, err)
			assert.Equal(t, "v2", data["k"])
		})
	})

	t.Run("Given a key that has been deleted", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return ErrSecretNotFound on Read", func(t *testing.T) {
			t.Parallel()
			suite := WhenStoringKvV1SecretBeforeEach()
			defer WhenStoringKvV1SecretAfterEach(suite)

			ctx := context.Background()
			assert.NoError(t, suite.sut.Write(ctx, "gamma", map[string]string{"k": "v"}))
			assert.NoError(t, suite.sut.Delete(ctx, "gamma"))

			data, err := suite.sut.Read(ctx, "gamma")
			assert.Nil(t, data)
			assert.True(t, errors.Is(err, runtime_common_interfaces.ErrSecretNotFound))
		})
	})
}
