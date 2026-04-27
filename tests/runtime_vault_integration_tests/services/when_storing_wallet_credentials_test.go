package services_test

import (
	"context"
	"testing"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	vault_services "github.com/6022-labs/agentic-agent-node/src/runtime_vault/services"
	vault_settings "github.com/6022-labs/agentic-agent-node/src/runtime_vault/settings"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/container_builders"
	vault "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
)

type WhenStoringWalletCredentialsTestingSuite struct {
	mountPath string
	sut       *vault_services.WalletCredentialsStorage
}

func WhenStoringWalletCredentialsBeforeEach() *WhenStoringWalletCredentialsTestingSuite {
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

	return &WhenStoringWalletCredentialsTestingSuite{
		mountPath: cfg.KvV1MountPath,
		sut:       vault_services.NewWalletCredentialsStorage(storage),
	}
}

func WhenStoringWalletCredentialsAfterEach(suite *WhenStoringWalletCredentialsTestingSuite) {
	container_builders.GetVaultContainerBuilder().DropMount(suite.mountPath)
}

func TestWhenStoringWalletCredentials(t *testing.T) {
	t.Parallel()

	t.Run("Given no wallet has been saved", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report not-found from Get", func(t *testing.T) {
			t.Parallel()
			suite := WhenStoringWalletCredentialsBeforeEach()
			defer WhenStoringWalletCredentialsAfterEach(suite)

			cred, found, err := suite.sut.Get(context.Background())
			assert.NoError(t, err)
			assert.False(t, found)
			assert.Nil(t, cred)
		})
	})

	t.Run("Given a wallet has been saved", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return the saved credentials from Get", func(t *testing.T) {
			t.Parallel()
			suite := WhenStoringWalletCredentialsBeforeEach()
			defer WhenStoringWalletCredentialsAfterEach(suite)

			ctx := context.Background()
			saved := &identity_interfaces.WalletCredentials{
				PrivateKeyHex: "0xb71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291",
				Address:       "0x14791697260E4c9A71f18484C9f997B308e59325",
			}
			assert.NoError(t, suite.sut.Save(ctx, saved))

			got, found, err := suite.sut.Get(ctx)
			assert.NoError(t, err)
			assert.True(t, found)
			assert.Equal(t, saved.PrivateKeyHex, got.PrivateKeyHex)
			assert.Equal(t, saved.Address, got.Address)
		})
	})

	t.Run("Given GetOrGenerate is called with no existing wallet", func(t *testing.T) {
		t.Parallel()

		t.Run("Should generate, persist, and return a fresh wallet", func(t *testing.T) {
			t.Parallel()
			suite := WhenStoringWalletCredentialsBeforeEach()
			defer WhenStoringWalletCredentialsAfterEach(suite)

			ctx := context.Background()
			cred, err := suite.sut.GetOrGenerate(ctx)
			assert.NoError(t, err)
			assert.NotNil(t, cred)
			assert.NotEmpty(t, cred.PrivateKeyHex)
			assert.NotEmpty(t, cred.Address)

			again, found, err := suite.sut.Get(ctx)
			assert.NoError(t, err)
			assert.True(t, found)
			assert.Equal(t, cred.PrivateKeyHex, again.PrivateKeyHex)
			assert.Equal(t, cred.Address, again.Address)
		})
	})

	t.Run("Given GetOrGenerate is called with an existing wallet", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return the existing wallet without regenerating", func(t *testing.T) {
			t.Parallel()
			suite := WhenStoringWalletCredentialsBeforeEach()
			defer WhenStoringWalletCredentialsAfterEach(suite)

			ctx := context.Background()
			first, err := suite.sut.GetOrGenerate(ctx)
			assert.NoError(t, err)
			second, err := suite.sut.GetOrGenerate(ctx)
			assert.NoError(t, err)
			assert.Equal(t, first.PrivateKeyHex, second.PrivateKeyHex)
			assert.Equal(t, first.Address, second.Address)
		})
	})
}
