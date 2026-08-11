package services_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	identity_services "github.com/6022-labs/agentic-agent-node/src/agent_identity/services"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	runtime_common_interfaces "github.com/6022-labs/agentic-agent-node/src/runtime_common/services/interfaces"
	identity_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	runtime_common_mocks "github.com/6022-labs/agentic-agent-node/tests/runtime_common_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenGettingWalletCredentialsForIdTestingSuite struct {
	sut *identity_services.WalletCredentialsStorage

	mockStorage *runtime_common_mocks.MockSecretStorageInterface
}

func WhenGettingWalletCredentialsForIdBeforeEach(t *testing.T) *WhenGettingWalletCredentialsForIdTestingSuite {
	mockController := gomock.NewController(t)

	mockStorage := runtime_common_mocks.NewMockSecretStorageInterface(mockController)
	mockAgentScopedKeyBuilder := identity_mocks.NewMockAgentScopedKeyBuilderInterface(mockController)
	mockAgentScopedKeyBuilder.EXPECT().ScopedKeyFor(gomock.Any(), gomock.Any()).DoAndReturn(
		func(agentId, key string) string { return "scope/" + key }).AnyTimes()
	mockGenerator := identity_mocks.NewMockAddressKeyGeneratorInterface(mockController)
	mockGenerator.EXPECT().Type().Return(enums.AddressTypeEvm).AnyTimes()

	return &WhenGettingWalletCredentialsForIdTestingSuite{
		sut: identity_services.NewWalletCredentialsStorage(identity_services.NewWalletCredentialsStorageParams{
			Storage:               mockStorage,
			AgentScopedKeyBuilder: mockAgentScopedKeyBuilder,
			Generators:            []identity_interfaces.AddressKeyGeneratorInterface{mockGenerator},
		}),
		mockStorage: mockStorage,
	}
}

func TestWhenGettingWalletCredentialsForId(t *testing.T) {
	t.Parallel()

	const scopedKey = "scope/wallet/evm"

	t.Run("Given no wallet has been persisted", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingWalletCredentialsForIdTestingSuite) {
			suite.mockStorage.EXPECT().
				Read(gomock.Any(), scopedKey).
				Return(nil, runtime_common_interfaces.ErrSecretNotFound)
		}

		t.Run("Should report not-found", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingWalletCredentialsForIdBeforeEach(t)
			initSuite(suite)

			cred, found, err := suite.sut.GetForId(context.Background(), "agent-1", enums.AddressTypeEvm)

			assert.NoError(t, err)
			assert.False(t, found)
			assert.Nil(t, cred)
		})
	})

	t.Run("Given a complete wallet has been persisted", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingWalletCredentialsForIdTestingSuite) {
			suite.mockStorage.EXPECT().Read(gomock.Any(), scopedKey).Return(map[string]string{
				"privateKey": "0xdeadbeef",
				"address":    "0xABCD",
			}, nil)
		}

		t.Run("Should return the credentials", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingWalletCredentialsForIdBeforeEach(t)
			initSuite(suite)

			cred, found, err := suite.sut.GetForId(context.Background(), "agent-1", enums.AddressTypeEvm)

			assert.NoError(t, err)
			assert.True(t, found)
			assert.Equal(t, "0xdeadbeef", cred.PrivateKeyHex)
			assert.Equal(t, "0xABCD", cred.Address)
		})
	})

	t.Run("Given the persisted secret is missing the privateKey field", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingWalletCredentialsForIdTestingSuite) {
			suite.mockStorage.EXPECT().Read(gomock.Any(), scopedKey).Return(map[string]string{
				"address": "0xABCD",
			}, nil)
		}

		t.Run("Should error", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingWalletCredentialsForIdBeforeEach(t)
			initSuite(suite)

			_, _, err := suite.sut.GetForId(context.Background(), "agent-1", enums.AddressTypeEvm)

			assert.Error(t, err)
		})
	})
}
