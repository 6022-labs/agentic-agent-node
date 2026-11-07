package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	identity_services "github.com/6022-labs/agentic-agent-node/src/agent_identity/services"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	identity_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	runtime_common_mocks "github.com/6022-labs/agentic-agent-node/tests/runtime_common_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenGettingOrGeneratingWalletCredentialsForIdTestingSuite struct {
	sut *identity_services.WalletCredentialsStorage

	mockStorage   *runtime_common_mocks.MockSecretStorageInterface
	mockGenerator *identity_mocks.MockAddressKeyGeneratorInterface
}

func WhenGettingOrGeneratingWalletCredentialsForIdBeforeEach(
	t *testing.T,
) *WhenGettingOrGeneratingWalletCredentialsForIdTestingSuite {
	mockController := gomock.NewController(t)

	mockStorage := runtime_common_mocks.NewMockSecretStorageInterface(mockController)
	mockAgentScopedKeyBuilder := identity_mocks.NewMockAgentScopedKeyBuilderInterface(mockController)
	mockAgentScopedKeyBuilder.EXPECT().ScopedKeyFor(gomock.Any(), gomock.Any()).DoAndReturn(
		func(agentId, key string) string { return "scope/" + key }).AnyTimes()
	mockGenerator := identity_mocks.NewMockAddressKeyGeneratorInterface(mockController)
	mockGenerator.EXPECT().Type().Return(enums.AddressTypeEvm).AnyTimes()

	return &WhenGettingOrGeneratingWalletCredentialsForIdTestingSuite{
		sut: identity_services.NewWalletCredentialsStorage(identity_services.NewWalletCredentialsStorageParams{
			Storage:               mockStorage,
			AgentScopedKeyBuilder: mockAgentScopedKeyBuilder,
			Generators:            []identity_interfaces.AddressKeyGeneratorInterface{mockGenerator},
		}),
		mockStorage:   mockStorage,
		mockGenerator: mockGenerator,
	}
}

func TestWhenGettingOrGeneratingWalletCredentialsForId(t *testing.T) {
	t.Parallel()

	const scopedKey = "scope/wallet/evm"

	t.Run("Given no generator is registered for the address type", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingOrGeneratingWalletCredentialsForIdTestingSuite) {
			suite.mockStorage.EXPECT().
				Read(gomock.Any(), "scope/wallet/btc").
				Return(nil, apperrors.ErrSecretNotFound)
		}

		t.Run("Should error without generating", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingOrGeneratingWalletCredentialsForIdBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.GetOrGenerateForId(context.Background(), "agent-1", enums.AddressTypeBtc)

			assert.ErrorContains(t, err, "no key generator")
		})
	})

	t.Run("Given saving the wallet fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingOrGeneratingWalletCredentialsForIdTestingSuite) {
			suite.mockStorage.EXPECT().
				Read(gomock.Any(), scopedKey).
				Return(nil, apperrors.ErrSecretNotFound)
			suite.mockGenerator.EXPECT().
				Generate().
				Return(&identity_interfaces.WalletCredentials{PrivateKeyHex: "0xnewpk", Address: "0xnewaddr"}, nil).
				Times(1)
			suite.mockStorage.EXPECT().Write(gomock.Any(), scopedKey, gomock.Any()).Return(errors.New("vault down"))
		}

		t.Run("Should propagate the persistence error", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingOrGeneratingWalletCredentialsForIdBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.GetOrGenerateForId(context.Background(), "agent-1", enums.AddressTypeEvm)

			assert.ErrorContains(t, err, "vault down")
		})
	})

	t.Run("Given a wallet already exists", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingOrGeneratingWalletCredentialsForIdTestingSuite) {
			suite.mockStorage.EXPECT().Read(gomock.Any(), scopedKey).Return(map[string]string{
				"privateKey": "0xpk",
				"address":    "0xaddr",
			}, nil)
		}

		t.Run("Should return it without generating or writing", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingOrGeneratingWalletCredentialsForIdBeforeEach(t)
			initSuite(suite)

			cred, err := suite.sut.GetOrGenerateForId(context.Background(), "agent-1", enums.AddressTypeEvm)

			assert.NoError(t, err)
			assert.Equal(t, "0xaddr", cred.Address)
		})
	})

	t.Run("Given no wallet exists", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingOrGeneratingWalletCredentialsForIdTestingSuite) {
			suite.mockStorage.EXPECT().
				Read(gomock.Any(), scopedKey).
				Return(nil, apperrors.ErrSecretNotFound)
			suite.mockGenerator.EXPECT().
				Generate().
				Return(&identity_interfaces.WalletCredentials{PrivateKeyHex: "0xnewpk", Address: "0xnewaddr"}, nil).
				Times(1)
			suite.mockStorage.EXPECT().Write(gomock.Any(), scopedKey, map[string]string{
				"privateKey": "0xnewpk",
				"address":    "0xnewaddr",
			}).Return(nil)
		}

		t.Run("Should generate a new wallet via the matching generator and persist it", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingOrGeneratingWalletCredentialsForIdBeforeEach(t)
			initSuite(suite)

			cred, err := suite.sut.GetOrGenerateForId(context.Background(), "agent-1", enums.AddressTypeEvm)

			assert.NoError(t, err)
			assert.Equal(t, "0xnewpk", cred.PrivateKeyHex)
			assert.Equal(t, "0xnewaddr", cred.Address)
		})
	})
}
