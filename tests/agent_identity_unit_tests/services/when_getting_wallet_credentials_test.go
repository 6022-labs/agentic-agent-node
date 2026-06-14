package services_test

import (
	"context"
	"testing"

	identity_services "github.com/6022-labs/agentic-agent-node/src/agent_identity/services"
	runtime_common_interfaces "github.com/6022-labs/agentic-agent-node/src/runtime_common/services/interfaces"
	identity_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	runtime_common_mocks "github.com/6022-labs/agentic-agent-node/tests/runtime_common_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenGettingWalletCredentialsTestingSuite struct {
	sut *identity_services.WalletCredentialsStorage

	mockStorage *runtime_common_mocks.MockSecretStorageInterface
}

func WhenGettingWalletCredentialsBeforeEach(t *testing.T) *WhenGettingWalletCredentialsTestingSuite {
	mockController := gomock.NewController(t)

	mockStorage := runtime_common_mocks.NewMockSecretStorageInterface(mockController)
	mockAgentScopedKeyBuilder := identity_mocks.NewMockAgentScopedKeyBuilderInterface(mockController)
	mockAgentScopedKeyBuilder.EXPECT().ScopedKey(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, key string) (string, error) { return "scope/" + key, nil }).AnyTimes()

	return &WhenGettingWalletCredentialsTestingSuite{
		sut:         identity_services.NewWalletCredentialsStorage(mockStorage, mockAgentScopedKeyBuilder),
		mockStorage: mockStorage,
	}
}

func TestWhenGettingWalletCredentials(t *testing.T) {
	t.Parallel()

	const scopedKey = "scope/wallet"

	t.Run("Given no wallet has been persisted", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingWalletCredentialsTestingSuite) {
			suite.mockStorage.EXPECT().Read(gomock.Any(), scopedKey).Return(nil, runtime_common_interfaces.ErrSecretNotFound)
		}

		t.Run("Should report not-found", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingWalletCredentialsBeforeEach(t)
			initSuite(suite)

			cred, found, err := suite.sut.Get(context.Background())

			assert.NoError(t, err)
			assert.False(t, found)
			assert.Nil(t, cred)
		})
	})

	t.Run("Given a complete wallet has been persisted", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingWalletCredentialsTestingSuite) {
			suite.mockStorage.EXPECT().Read(gomock.Any(), scopedKey).Return(map[string]string{
				"privateKey": "0xdeadbeef",
				"address":    "0xABCD",
			}, nil)
		}

		t.Run("Should return the credentials", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingWalletCredentialsBeforeEach(t)
			initSuite(suite)

			cred, found, err := suite.sut.Get(context.Background())

			assert.NoError(t, err)
			assert.True(t, found)
			assert.Equal(t, "0xdeadbeef", cred.PrivateKeyHex)
			assert.Equal(t, "0xABCD", cred.Address)
		})
	})

	t.Run("Given the persisted secret is missing the privateKey field", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingWalletCredentialsTestingSuite) {
			suite.mockStorage.EXPECT().Read(gomock.Any(), scopedKey).Return(map[string]string{
				"address": "0xABCD",
			}, nil)
		}

		t.Run("Should error", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingWalletCredentialsBeforeEach(t)
			initSuite(suite)

			_, _, err := suite.sut.Get(context.Background())

			assert.Error(t, err)
		})
	})
}
