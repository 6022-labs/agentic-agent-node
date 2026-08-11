package services_test

import (
	"context"
	"errors"
	"testing"

	runtime_common_interfaces "github.com/6022-labs/agentic-agent-node/src/runtime_common/services/interfaces"
	vault_services "github.com/6022-labs/agentic-agent-node/src/runtime_vault/services"
	vault_settings "github.com/6022-labs/agentic-agent-node/src/runtime_vault/settings"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/container_builders"
	metrics_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/runtime_vault_mocks/metrics_mocks/interfaces_mocks"
	vault "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenReadingKvV1SecretTestingSuite struct {
	mountPath string
	sut       *vault_services.KvV1SecretStorage
}

func WhenReadingKvV1SecretBeforeEach(t *testing.T) *WhenReadingKvV1SecretTestingSuite {
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

	return &WhenReadingKvV1SecretTestingSuite{
		mountPath: cfg.KvV1MountPath,
		sut:       vault_services.NewKvV1SecretStorage(client, settings, mockMetrics),
	}
}

func WhenReadingKvV1SecretAfterEach(suite *WhenReadingKvV1SecretTestingSuite) {
	container_builders.GetVaultContainerBuilder().DropMount(suite.mountPath)
}

func TestWhenReadingKvV1Secret(t *testing.T) {
	t.Parallel()

	t.Run("Given a key that has not been written", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return ErrSecretNotFound", func(t *testing.T) {
			t.Parallel()
			suite := WhenReadingKvV1SecretBeforeEach(t)
			defer WhenReadingKvV1SecretAfterEach(suite)

			data, err := suite.sut.Read(context.Background(), "missing")
			assert.Nil(t, data)
			assert.True(t, errors.Is(err, runtime_common_interfaces.ErrSecretNotFound))
		})
	})
}
