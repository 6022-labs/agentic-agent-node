package factory_test

import (
	"context"
	"testing"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/factory"
	repository_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	interface_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// A fixed secp256k1 test key and its derived address.
const testPrivateKeyHex = "4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318"

type WhenCreatingTransactorTestingSuite struct {
	sut *factory.TransactorFactory

	mockWalletCredentialsStorage *interface_mocks.MockWalletCredentialsStorageInterface
	mockAgentRepository          *repository_mocks.MockAgentRepositoryInterface
}

func WhenCreatingTransactorBeforeEach(t *testing.T) *WhenCreatingTransactorTestingSuite {
	mockController := gomock.NewController(t)

	mockWalletCredentialsStorage := interface_mocks.NewMockWalletCredentialsStorageInterface(mockController)
	mockAgentRepository := repository_mocks.NewMockAgentRepositoryInterface(mockController)
	mockAgentRepository.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil).AnyTimes()

	sut := factory.NewTransactorFactory(mockWalletCredentialsStorage, mockAgentRepository)

	return &WhenCreatingTransactorTestingSuite{
		sut:                          sut,
		mockWalletCredentialsStorage: mockWalletCredentialsStorage,
		mockAgentRepository:          mockAgentRepository,
	}
}

func TestWhenCreatingTransactor(t *testing.T) {
	t.Parallel()

	t.Run("Given the node wallet is available", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingTransactorTestingSuite) {
			suite.mockWalletCredentialsStorage.EXPECT().GetForId(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&identity_interfaces.WalletCredentials{PrivateKeyHex: testPrivateKeyHex}, true, nil)
		}

		t.Run("Should build a transactor signed by that wallet", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingTransactorBeforeEach(t)
			initSuite(suite)

			opts, err := suite.sut.Create(context.Background(), 80002)

			assert.NoError(t, err)
			key, _ := crypto.HexToECDSA(testPrivateKeyHex)
			assert.Equal(t, crypto.PubkeyToAddress(key.PublicKey).Hex(), opts.From.Hex())
		})
	})

	t.Run("Given the node wallet is not bootstrapped", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingTransactorTestingSuite) {
			suite.mockWalletCredentialsStorage.EXPECT().
				GetForId(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, false, nil)
		}

		t.Run("Should return an error", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingTransactorBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Create(context.Background(), 1)

			assert.ErrorContains(t, err, "wallet credentials not present")
		})
	})
}
