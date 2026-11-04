package factory_test

import (
	"context"
	"testing"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment_blockchain/factory"
	repository_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	interface_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// A fixed secp256k1 test key and its derived address.
const testPrivateKeyHex = "4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318"

type WhenCreatingATransactorTestingSuite struct {
	sut *factory.TransactorFactory

	mockWalletCredentialsStorage *interface_mocks.MockWalletCredentialsStorageInterface
	mockAgentRepository          *repository_mocks.MockAgentRepositoryInterface
}

func WhenCreatingATransactorBeforeEach(t *testing.T) *WhenCreatingATransactorTestingSuite {
	mockController := gomock.NewController(t)

	mockWalletCredentialsStorage := interface_mocks.NewMockWalletCredentialsStorageInterface(mockController)
	mockAgentRepository := repository_mocks.NewMockAgentRepositoryInterface(mockController)

	return &WhenCreatingATransactorTestingSuite{
		sut:                          factory.NewTransactorFactory(mockWalletCredentialsStorage, mockAgentRepository),
		mockWalletCredentialsStorage: mockWalletCredentialsStorage,
		mockAgentRepository:          mockAgentRepository,
	}
}

func TestWhenCreatingATransactor(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given loading the agent fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingATransactorTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(ctx).Return(nil, assert.AnError).Times(1)
		}

		t.Run("Should return the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingATransactorBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Create(ctx, 1)

			assert.ErrorIs(t, err, assert.AnError)
		})
	})

	t.Run("Given the agent is not bootstrapped", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingATransactorTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(ctx).Return(nil, nil).Times(1)
		}

		t.Run("Should return an error", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingATransactorBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Create(ctx, 1)

			assert.ErrorContains(t, err, "agent not bootstrapped")
		})
	})

	t.Run("Given loading the wallet credentials fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingATransactorTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(ctx).Return(model_builders.NewAgentBuilder().Build(), nil).Times(1)
			suite.mockWalletCredentialsStorage.EXPECT().
				GetForId(ctx, gomock.Any(), gomock.Any()).
				Return(nil, false, assert.AnError).
				Times(1)
		}

		t.Run("Should return the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingATransactorBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Create(ctx, 1)

			assert.ErrorIs(t, err, assert.AnError)
		})
	})

	t.Run("Given the node wallet is not bootstrapped", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingATransactorTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(ctx).Return(model_builders.NewAgentBuilder().Build(), nil).Times(1)
			suite.mockWalletCredentialsStorage.EXPECT().
				GetForId(ctx, gomock.Any(), gomock.Any()).
				Return(nil, false, nil).
				Times(1)
		}

		t.Run("Should return an error", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingATransactorBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Create(ctx, 1)

			assert.ErrorContains(t, err, "wallet credentials not present")
		})
	})

	t.Run("Given the node wallet key is invalid", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingATransactorTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(ctx).Return(model_builders.NewAgentBuilder().Build(), nil).Times(1)
			suite.mockWalletCredentialsStorage.EXPECT().
				GetForId(ctx, gomock.Any(), gomock.Any()).
				Return(&identity_interfaces.WalletCredentials{PrivateKeyHex: "not-a-key"}, true, nil).
				Times(1)
		}

		t.Run("Should return an error", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingATransactorBeforeEach(t)
			initSuite(suite)

			opts, err := suite.sut.Create(ctx, 1)

			assert.ErrorContains(t, err, "invalid wallet private key")
			assert.Nil(t, opts)
		})
	})

	t.Run("Given the node wallet is available", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingATransactorTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(ctx).Return(model_builders.NewAgentBuilder().Build(), nil).Times(1)
			suite.mockWalletCredentialsStorage.EXPECT().
				GetForId(ctx, gomock.Any(), gomock.Any()).
				Return(&identity_interfaces.WalletCredentials{PrivateKeyHex: testPrivateKeyHex}, true, nil).
				Times(1)
		}

		t.Run("Should build a transactor signed by that wallet", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingATransactorBeforeEach(t)
			initSuite(suite)

			opts, err := suite.sut.Create(ctx, 80002)

			assert.NoError(t, err)
			key, _ := crypto.HexToECDSA(testPrivateKeyHex)
			assert.Equal(t, crypto.PubkeyToAddress(key.PublicKey).Hex(), opts.From.Hex())
		})
	})
}
