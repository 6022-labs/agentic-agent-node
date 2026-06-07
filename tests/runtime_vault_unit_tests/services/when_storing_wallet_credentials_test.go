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

type WhenStoringWalletCredentialsTestingSuite struct {
	sut *vault_services.WalletCredentialsStorage

	mockStorage *vault_mocks.MockKvV1SecretStorageInterface
}

func WhenStoringWalletCredentialsBeforeEach(t *testing.T) *WhenStoringWalletCredentialsTestingSuite {
	mockController := gomock.NewController(t)

	mockStorage := vault_mocks.NewMockKvV1SecretStorageInterface(mockController)
	sut := vault_services.NewWalletCredentialsStorage(mockStorage)

	return &WhenStoringWalletCredentialsTestingSuite{
		sut: sut,

		mockStorage: mockStorage,
	}
}

func TestWhenStoringWalletCredentials(t *testing.T) {
	t.Parallel()

	t.Run("Given the underlying storage has no wallet yet", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenStoringWalletCredentialsTestingSuite) {
			suite.mockStorage.EXPECT().Read(gomock.Any(), walletKey).Return(nil, vault_services.ErrSecretNotFound)
		}

		t.Run("Should return found=false with no error", func(t *testing.T) {
			t.Parallel()

			suite := WhenStoringWalletCredentialsBeforeEach(t)
			initSuite(suite)

			cred, found, err := suite.sut.Get(context.Background())

			assert.NoError(t, err)
			assert.False(t, found)
			assert.Nil(t, cred)
		})
	})

	t.Run("Given the underlying storage holds a complete wallet", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenStoringWalletCredentialsTestingSuite) {
			suite.mockStorage.EXPECT().Read(gomock.Any(), walletKey).Return(map[string]string{
				"privateKey": "0xdeadbeef",
				"address":    "0xABCD",
			}, nil)
		}

		t.Run("Should return the credentials", func(t *testing.T) {
			t.Parallel()

			suite := WhenStoringWalletCredentialsBeforeEach(t)
			initSuite(suite)

			cred, found, err := suite.sut.Get(context.Background())

			assert.NoError(t, err)
			assert.True(t, found)
			assert.Equal(t, "0xdeadbeef", cred.PrivateKeyHex)
			assert.Equal(t, "0xABCD", cred.Address)
		})
	})

	t.Run("Given the stored secret is missing the privateKey field", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenStoringWalletCredentialsTestingSuite) {
			suite.mockStorage.EXPECT().Read(gomock.Any(), walletKey).Return(map[string]string{
				"address": "0xABCD",
			}, nil)
		}

		t.Run("Should error", func(t *testing.T) {
			t.Parallel()

			suite := WhenStoringWalletCredentialsBeforeEach(t)
			initSuite(suite)

			_, _, err := suite.sut.Get(context.Background())

			assert.Error(t, err)
		})
	})

	t.Run("Given Save is called with valid credentials", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenStoringWalletCredentialsTestingSuite) {
			suite.mockStorage.EXPECT().Write(gomock.Any(), walletKey, map[string]string{
				"privateKey": "0xpk",
				"address":    "0xaddr",
			}).Return(nil)
		}

		t.Run("Should write to storage", func(t *testing.T) {
			t.Parallel()

			suite := WhenStoringWalletCredentialsBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Save(context.Background(), &identity_interfaces.WalletCredentials{PrivateKeyHex: "0xpk", Address: "0xaddr"})

			assert.NoError(t, err)
		})
	})

	t.Run("Given Save is called with empty credentials", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenStoringWalletCredentialsTestingSuite) {}

		t.Run("Should reject without writing", func(t *testing.T) {
			t.Parallel()

			suite := WhenStoringWalletCredentialsBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Save(context.Background(), &identity_interfaces.WalletCredentials{})

			assert.Error(t, err)
		})
	})

	t.Run("Given GetOrGenerate finds an existing wallet", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenStoringWalletCredentialsTestingSuite) {
			suite.mockStorage.EXPECT().Read(gomock.Any(), walletKey).Return(map[string]string{
				"privateKey": "0xpk",
				"address":    "0xaddr",
			}, nil)
		}

		t.Run("Should return it without writing", func(t *testing.T) {
			t.Parallel()

			suite := WhenStoringWalletCredentialsBeforeEach(t)
			initSuite(suite)

			cred, err := suite.sut.GetOrGenerate(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, "0xaddr", cred.Address)
		})
	})

	t.Run("Given GetOrGenerate finds no wallet", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenStoringWalletCredentialsTestingSuite) {
			suite.mockStorage.EXPECT().Read(gomock.Any(), walletKey).Return(nil, vault_services.ErrSecretNotFound)
			suite.mockStorage.EXPECT().Write(gomock.Any(), walletKey, gomock.Any()).Return(nil)
		}

		t.Run("Should generate a new wallet and persist it", func(t *testing.T) {
			t.Parallel()

			suite := WhenStoringWalletCredentialsBeforeEach(t)
			initSuite(suite)

			cred, err := suite.sut.GetOrGenerate(context.Background())

			assert.NoError(t, err)
			assert.NotEmpty(t, cred.PrivateKeyHex)
			assert.NotEmpty(t, cred.Address)
		})
	})

	t.Run("Given GetOrGenerate fails to persist the new wallet", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenStoringWalletCredentialsTestingSuite) {
			suite.mockStorage.EXPECT().Read(gomock.Any(), walletKey).Return(nil, vault_services.ErrSecretNotFound)
			suite.mockStorage.EXPECT().Write(gomock.Any(), walletKey, gomock.Any()).Return(errors.New("vault down"))
		}

		t.Run("Should propagate the persistence error", func(t *testing.T) {
			t.Parallel()

			suite := WhenStoringWalletCredentialsBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.GetOrGenerate(context.Background())

			assert.ErrorContains(t, err, "vault down")
		})
	})
}
