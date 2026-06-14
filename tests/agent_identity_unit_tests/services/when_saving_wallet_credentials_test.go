package services_test

import (
	"context"
	"testing"

	identity_services "github.com/6022-labs/agentic-agent-node/src/agent_identity/services"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	identity_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	runtime_common_mocks "github.com/6022-labs/agentic-agent-node/tests/runtime_common_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenSavingWalletCredentialsTestingSuite struct {
	sut *identity_services.WalletCredentialsStorage

	mockStorage *runtime_common_mocks.MockSecretStorageInterface
}

func WhenSavingWalletCredentialsBeforeEach(t *testing.T) *WhenSavingWalletCredentialsTestingSuite {
	mockController := gomock.NewController(t)

	mockStorage := runtime_common_mocks.NewMockSecretStorageInterface(mockController)
	mockAgentScopedKeyBuilder := identity_mocks.NewMockAgentScopedKeyBuilderInterface(mockController)
	mockAgentScopedKeyBuilder.EXPECT().ScopedKey(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, key string) (string, error) { return "scope/" + key, nil }).AnyTimes()

	return &WhenSavingWalletCredentialsTestingSuite{
		sut:         identity_services.NewWalletCredentialsStorage(mockStorage, mockAgentScopedKeyBuilder),
		mockStorage: mockStorage,
	}
}

func TestWhenSavingWalletCredentials(t *testing.T) {
	t.Parallel()

	const scopedKey = "scope/wallet"

	t.Run("Given valid credentials", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSavingWalletCredentialsTestingSuite) {
			suite.mockStorage.EXPECT().Write(gomock.Any(), scopedKey, map[string]string{
				"privateKey": "0xpk",
				"address":    "0xaddr",
			}).Return(nil)
		}

		t.Run("Should write them to the scoped key", func(t *testing.T) {
			t.Parallel()

			suite := WhenSavingWalletCredentialsBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Save(context.Background(), &identity_interfaces.WalletCredentials{PrivateKeyHex: "0xpk", Address: "0xaddr"})

			assert.NoError(t, err)
		})
	})

	t.Run("Given empty credentials", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject without writing", func(t *testing.T) {
			t.Parallel()

			suite := WhenSavingWalletCredentialsBeforeEach(t)

			err := suite.sut.Save(context.Background(), &identity_interfaces.WalletCredentials{})

			assert.Error(t, err)
		})
	})
}
