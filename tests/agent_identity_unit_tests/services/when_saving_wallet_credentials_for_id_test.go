package services_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	identity_services "github.com/6022-labs/agentic-agent-node/src/agent_identity/services"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	identity_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	runtime_common_mocks "github.com/6022-labs/agentic-agent-node/tests/runtime_common_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenSavingWalletCredentialsForIdTestingSuite struct {
	sut *identity_services.WalletCredentialsStorage

	mockStorage *runtime_common_mocks.MockSecretStorageInterface
}

func WhenSavingWalletCredentialsForIdBeforeEach(t *testing.T) *WhenSavingWalletCredentialsForIdTestingSuite {
	mockController := gomock.NewController(t)

	mockStorage := runtime_common_mocks.NewMockSecretStorageInterface(mockController)
	mockAgentScopedKeyBuilder := identity_mocks.NewMockAgentScopedKeyBuilderInterface(mockController)
	mockAgentScopedKeyBuilder.EXPECT().ScopedKeyFor(gomock.Any(), gomock.Any()).DoAndReturn(
		func(agentId, key string) string { return "scope/" + key }).AnyTimes()
	mockGenerator := identity_mocks.NewMockAddressKeyGeneratorInterface(mockController)
	mockGenerator.EXPECT().Type().Return(enums.AddressTypeEvm).AnyTimes()

	return &WhenSavingWalletCredentialsForIdTestingSuite{
		sut: identity_services.NewWalletCredentialsStorage(identity_services.NewWalletCredentialsStorageParams{
			Storage:               mockStorage,
			AgentScopedKeyBuilder: mockAgentScopedKeyBuilder,
			Generators:            []identity_interfaces.AddressKeyGeneratorInterface{mockGenerator},
		}),
		mockStorage: mockStorage,
	}
}

func TestWhenSavingWalletCredentialsForId(t *testing.T) {
	t.Parallel()

	const scopedKey = "scope/wallet/evm"

	t.Run("Given valid credentials", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSavingWalletCredentialsForIdTestingSuite) {
			suite.mockStorage.EXPECT().Write(gomock.Any(), scopedKey, map[string]string{
				"privateKey": "0xpk",
				"address":    "0xaddr",
			}).Return(nil)
		}

		t.Run("Should write them to the scoped key", func(t *testing.T) {
			t.Parallel()

			suite := WhenSavingWalletCredentialsForIdBeforeEach(t)
			initSuite(suite)

			err := suite.sut.SaveForId(context.Background(), "agent-1", enums.AddressTypeEvm, &identity_interfaces.WalletCredentials{PrivateKeyHex: "0xpk", Address: "0xaddr"})

			assert.NoError(t, err)
		})
	})

	t.Run("Given empty credentials", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject without writing", func(t *testing.T) {
			t.Parallel()

			suite := WhenSavingWalletCredentialsForIdBeforeEach(t)

			err := suite.sut.SaveForId(context.Background(), "agent-1", enums.AddressTypeEvm, &identity_interfaces.WalletCredentials{})

			assert.Error(t, err)
		})
	})
}
