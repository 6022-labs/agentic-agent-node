package services_test

import (
	"context"
	"testing"

	vault_services "github.com/6022-labs/agentic-agent-node/src/runtime_vault/services"
	vault_settings "github.com/6022-labs/agentic-agent-node/src/runtime_vault/settings"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/container_builders"
	metrics_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/runtime_vault_mocks/metrics_mocks/interfaces_mocks"
	vault "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenWritingKvV1SecretTestingSuite struct {
	mountPath string
	sut       *vault_services.KvV1SecretStorage
}

func WhenWritingKvV1SecretBeforeEach(t *testing.T) *WhenWritingKvV1SecretTestingSuite {
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

	mockMetrics := metrics_interfaces_mocks.NewMockVaultMetricsInterface(gomock.NewController(t))
	mockMetrics.EXPECT().
		RecordOperation(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes()

	return &WhenWritingKvV1SecretTestingSuite{
		mountPath: cfg.KvV1MountPath,
		sut:       vault_services.NewKvV1SecretStorage(client, settings, mockMetrics),
	}
}

func WhenWritingKvV1SecretAfterEach(suite *WhenWritingKvV1SecretTestingSuite) {
	container_builders.GetVaultContainerBuilder().DropMount(suite.mountPath)
}

func TestWhenWritingKvV1Secret(t *testing.T) {
	t.Parallel()

	t.Run("Given a new key", func(t *testing.T) {
		t.Parallel()

		t.Run("Should persist the values retrievable by Read", func(t *testing.T) {
			t.Parallel()
			suite := WhenWritingKvV1SecretBeforeEach(t)
			defer WhenWritingKvV1SecretAfterEach(suite)

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

	t.Run("Given a key exists with updated values", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return the latest values on Read", func(t *testing.T) {
			t.Parallel()
			suite := WhenWritingKvV1SecretBeforeEach(t)
			defer WhenWritingKvV1SecretAfterEach(suite)

			ctx := context.Background()
			assert.NoError(t, suite.sut.Write(ctx, "beta", map[string]string{"k": "v1"}))
			assert.NoError(t, suite.sut.Write(ctx, "beta", map[string]string{"k": "v2"}))

			data, err := suite.sut.Read(ctx, "beta")
			assert.NoError(t, err)
			assert.Equal(t, "v2", data["k"])
		})
	})
}
