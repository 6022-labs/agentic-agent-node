package services_test

import (
	"context"
	"errors"
	"testing"

	identity_services "github.com/6022-labs/agentic-agent-node/src/agent_identity/services"
	runtime_common_interfaces "github.com/6022-labs/agentic-agent-node/src/runtime_common/services/interfaces"
	identity_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	runtime_common_mocks "github.com/6022-labs/agentic-agent-node/tests/runtime_common_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenGettingOrGeneratingWalletCredentialsTestingSuite struct {
	sut *identity_services.WalletCredentialsStorage

	mockStorage *runtime_common_mocks.MockSecretStorageInterface
}

func WhenGettingOrGeneratingWalletCredentialsBeforeEach(t *testing.T) *WhenGettingOrGeneratingWalletCredentialsTestingSuite {
	mockController := gomock.NewController(t)

	mockStorage := runtime_common_mocks.NewMockSecretStorageInterface(mockController)
	mockAgentScopedKeyBuilder := identity_mocks.NewMockAgentScopedKeyBuilderInterface(mockController)
	mockAgentScopedKeyBuilder.EXPECT().ScopedKey(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, key string) (string, error) { return "scope/" + key, nil }).AnyTimes()

	return &WhenGettingOrGeneratingWalletCredentialsTestingSuite{
		sut:         identity_services.NewWalletCredentialsStorage(mockStorage, mockAgentScopedKeyBuilder),
		mockStorage: mockStorage,
	}
}

func TestWhenGettingOrGeneratingWalletCredentials(t *testing.T) {
	t.Parallel()

	const scopedKey = "scope/wallet"

	t.Run("Given a wallet already exists", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingOrGeneratingWalletCredentialsTestingSuite) {
			suite.mockStorage.EXPECT().Read(gomock.Any(), scopedKey).Return(map[string]string{
				"privateKey": "0xpk",
				"address":    "0xaddr",
			}, nil)
		}

		t.Run("Should return it without writing", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingOrGeneratingWalletCredentialsBeforeEach(t)
			initSuite(suite)

			cred, err := suite.sut.GetOrGenerate(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, "0xaddr", cred.Address)
		})
	})

	t.Run("Given no wallet exists yet", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingOrGeneratingWalletCredentialsTestingSuite) {
			suite.mockStorage.EXPECT().Read(gomock.Any(), scopedKey).Return(nil, runtime_common_interfaces.ErrSecretNotFound)
			suite.mockStorage.EXPECT().Write(gomock.Any(), scopedKey, gomock.Any()).Return(nil)
		}

		t.Run("Should generate a new wallet and persist it", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingOrGeneratingWalletCredentialsBeforeEach(t)
			initSuite(suite)

			cred, err := suite.sut.GetOrGenerate(context.Background())

			assert.NoError(t, err)
			assert.NotEmpty(t, cred.PrivateKeyHex)
			assert.NotEmpty(t, cred.Address)
		})
	})

	t.Run("Given persisting the generated wallet fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingOrGeneratingWalletCredentialsTestingSuite) {
			suite.mockStorage.EXPECT().Read(gomock.Any(), scopedKey).Return(nil, runtime_common_interfaces.ErrSecretNotFound)
			suite.mockStorage.EXPECT().Write(gomock.Any(), scopedKey, gomock.Any()).Return(errors.New("vault down"))
		}

		t.Run("Should propagate the persistence error", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingOrGeneratingWalletCredentialsBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.GetOrGenerate(context.Background())

			assert.ErrorContains(t, err, "vault down")
		})
	})
}
