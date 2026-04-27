package services_test

import (
	"context"
	"errors"
	"testing"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	vault_services "github.com/6022-labs/agentic-agent-node/src/runtime_vault/services"
	vault_mocks "github.com/6022-labs/agentic-agent-node/tests/runtime_vault_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

const walletKey = "agent/wallet"

func TestWhenStoringWalletCredentials(t *testing.T) {
	t.Parallel()

	t.Run("Given the underlying storage has no wallet yet", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		storage := vault_mocks.NewMockKvV1SecretStorageInterface(ctrl)
		storage.EXPECT().Read(gomock.Any(), walletKey).Return(nil, vault_services.ErrSecretNotFound)
		wallet := vault_services.NewWalletCredentialsStorage(storage)

		t.Run("Get should return found=false with no error", func(t *testing.T) {
			cred, found, err := wallet.Get(context.Background())
			assert.NoError(t, err)
			assert.False(t, found)
			assert.Nil(t, cred)
		})
	})

	t.Run("Given the underlying storage holds a complete wallet", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		storage := vault_mocks.NewMockKvV1SecretStorageInterface(ctrl)
		storage.EXPECT().Read(gomock.Any(), walletKey).Return(map[string]string{
			"privateKey": "0xdeadbeef",
			"address":    "0xABCD",
		}, nil)
		wallet := vault_services.NewWalletCredentialsStorage(storage)

		t.Run("Get should return the credentials", func(t *testing.T) {
			cred, found, err := wallet.Get(context.Background())
			assert.NoError(t, err)
			assert.True(t, found)
			assert.Equal(t, "0xdeadbeef", cred.PrivateKeyHex)
			assert.Equal(t, "0xABCD", cred.Address)
		})
	})

	t.Run("Given the stored secret is missing the privateKey field", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		storage := vault_mocks.NewMockKvV1SecretStorageInterface(ctrl)
		storage.EXPECT().Read(gomock.Any(), walletKey).Return(map[string]string{
			"address": "0xABCD",
		}, nil)
		wallet := vault_services.NewWalletCredentialsStorage(storage)

		t.Run("Get should error", func(t *testing.T) {
			_, _, err := wallet.Get(context.Background())
			assert.Error(t, err)
		})
	})

	t.Run("Given Save is called with valid credentials", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		storage := vault_mocks.NewMockKvV1SecretStorageInterface(ctrl)
		storage.EXPECT().Write(gomock.Any(), walletKey, map[string]string{
			"privateKey": "0xpk",
			"address":    "0xaddr",
		}).Return(nil)
		wallet := vault_services.NewWalletCredentialsStorage(storage)

		t.Run("Should write to storage", func(t *testing.T) {
			err := wallet.Save(context.Background(), &identity_interfaces.WalletCredentials{PrivateKeyHex: "0xpk", Address: "0xaddr"})
			assert.NoError(t, err)
		})
	})

	t.Run("Given Save is called with empty credentials", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		storage := vault_mocks.NewMockKvV1SecretStorageInterface(ctrl)
		wallet := vault_services.NewWalletCredentialsStorage(storage)

		t.Run("Should reject without writing", func(t *testing.T) {
			err := wallet.Save(context.Background(), &identity_interfaces.WalletCredentials{})
			assert.Error(t, err)
		})
	})

	t.Run("Given GetOrGenerate finds an existing wallet", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		storage := vault_mocks.NewMockKvV1SecretStorageInterface(ctrl)
		storage.EXPECT().Read(gomock.Any(), walletKey).Return(map[string]string{
			"privateKey": "0xpk",
			"address":    "0xaddr",
		}, nil)
		wallet := vault_services.NewWalletCredentialsStorage(storage)

		t.Run("Should return it without writing", func(t *testing.T) {
			cred, err := wallet.GetOrGenerate(context.Background())
			assert.NoError(t, err)
			assert.Equal(t, "0xaddr", cred.Address)
		})
	})

	t.Run("Given GetOrGenerate finds no wallet", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		storage := vault_mocks.NewMockKvV1SecretStorageInterface(ctrl)
		storage.EXPECT().Read(gomock.Any(), walletKey).Return(nil, vault_services.ErrSecretNotFound)
		storage.EXPECT().Write(gomock.Any(), walletKey, gomock.Any()).Return(nil)
		wallet := vault_services.NewWalletCredentialsStorage(storage)

		t.Run("Should generate a new wallet and persist it", func(t *testing.T) {
			cred, err := wallet.GetOrGenerate(context.Background())
			assert.NoError(t, err)
			assert.NotEmpty(t, cred.PrivateKeyHex)
			assert.NotEmpty(t, cred.Address)
		})
	})

	t.Run("Given GetOrGenerate fails to persist the new wallet", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		storage := vault_mocks.NewMockKvV1SecretStorageInterface(ctrl)
		storage.EXPECT().Read(gomock.Any(), walletKey).Return(nil, vault_services.ErrSecretNotFound)
		storage.EXPECT().Write(gomock.Any(), walletKey, gomock.Any()).Return(errors.New("vault down"))
		wallet := vault_services.NewWalletCredentialsStorage(storage)

		t.Run("Should propagate the persistence error", func(t *testing.T) {
			_, err := wallet.GetOrGenerate(context.Background())
			assert.ErrorContains(t, err, "vault down")
		})
	})
}
