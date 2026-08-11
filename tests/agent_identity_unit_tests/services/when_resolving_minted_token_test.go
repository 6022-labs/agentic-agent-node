package services_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/services"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	repository_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	interface_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenResolvingMintedTokenTestingSuite struct {
	sut *services.MintPublisher

	mockAgentCollectionRequester         *interface_mocks.MockAgentCollectionRequesterInterface
	mockAgentCollectionsManagerRequester *interface_mocks.MockAgentCollectionsManagerRequesterInterface
	mockAgentMinter                      *interface_mocks.MockAgentMinterInterface
	mockWalletCredentialsStorage         *interface_mocks.MockWalletCredentialsStorageInterface
	mockAgentRepository                  *repository_mocks.MockAgentRepositoryInterface
	mockAgentAddressRepository           *repository_mocks.MockAgentAddressRepositoryInterface
	mockAgentRegistrationStateRepository *repository_mocks.MockAgentRegistrationStateRepositoryInterface
	mockRegistrationReconciler           *interface_mocks.MockRegistrationReconcilerInterface
}

func WhenResolvingMintedTokenBeforeEach(t *testing.T) *WhenResolvingMintedTokenTestingSuite {
	mockController := gomock.NewController(t)

	mockAgentCollectionRequester := interface_mocks.NewMockAgentCollectionRequesterInterface(mockController)
	mockAgentCollectionsManagerRequester := interface_mocks.NewMockAgentCollectionsManagerRequesterInterface(
		mockController,
	)
	mockAgentMinter := interface_mocks.NewMockAgentMinterInterface(mockController)
	mockWalletCredentialsStorage := interface_mocks.NewMockWalletCredentialsStorageInterface(mockController)
	mockAgentRepository := repository_mocks.NewMockAgentRepositoryInterface(mockController)
	mockAgentAddressRepository := repository_mocks.NewMockAgentAddressRepositoryInterface(mockController)
	mockAgentRegistrationStateRepository := repository_mocks.NewMockAgentRegistrationStateRepositoryInterface(
		mockController,
	)
	mockRegistrationReconciler := interface_mocks.NewMockRegistrationReconcilerInterface(mockController)

	sut := services.NewMintPublisher(
		mockAgentCollectionRequester,
		mockAgentCollectionsManagerRequester,
		mockAgentMinter,
		mockWalletCredentialsStorage,
		mockAgentRepository,
		mockAgentAddressRepository,
		mockAgentRegistrationStateRepository,
		mockRegistrationReconciler,
	)

	return &WhenResolvingMintedTokenTestingSuite{
		sut:                                  sut,
		mockAgentCollectionRequester:         mockAgentCollectionRequester,
		mockAgentCollectionsManagerRequester: mockAgentCollectionsManagerRequester,
		mockAgentMinter:                      mockAgentMinter,
		mockWalletCredentialsStorage:         mockWalletCredentialsStorage,
		mockAgentRepository:                  mockAgentRepository,
		mockAgentAddressRepository:           mockAgentAddressRepository,
		mockAgentRegistrationStateRepository: mockAgentRegistrationStateRepository,
		mockRegistrationReconciler:           mockRegistrationReconciler,
	}
}

func TestWhenResolvingMintedToken(t *testing.T) {
	t.Parallel()

	nodeWallet := common_tests.RandomEthAddress().Hex()
	agent := model_builders.NewAgentBuilder().WithoutCollectionTokenId().Build()

	t.Run("Given the node wallet controls a token in the draft collection", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenResolvingMintedTokenTestingSuite) {
			suite.mockWalletCredentialsStorage.EXPECT().
				GetForId(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&identity_interfaces.WalletCredentials{Address: nodeWallet}, true, nil)
			suite.mockAgentCollectionRequester.EXPECT().
				AddressToTokenId(gomock.Any(), gomock.Any(), enums.AddressTypeEvm, gomock.Any()).
				Return(uint64(7), nil)
		}

		t.Run("Should return that token without scanning the manager", func(t *testing.T) {
			t.Parallel()

			suite := WhenResolvingMintedTokenBeforeEach(t)
			initSuite(suite)

			_, tokenId, err := suite.sut.ResolveMintedToken(context.Background(), agent)

			assert.NoError(t, err)
			assert.Equal(t, uint64(7), tokenId)
		})
	})

	t.Run("Given the draft collection has no token but the manager finds one", func(t *testing.T) {
		t.Parallel()

		foundCollection := common_tests.RandomEthAddress()

		initSuite := func(suite *WhenResolvingMintedTokenTestingSuite) {
			suite.mockWalletCredentialsStorage.EXPECT().
				GetForId(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&identity_interfaces.WalletCredentials{Address: nodeWallet}, true, nil)
			suite.mockAgentCollectionRequester.EXPECT().
				AddressToTokenId(gomock.Any(), gomock.Any(), enums.AddressTypeEvm, gomock.Any()).
				Return(uint64(0), nil)
			suite.mockAgentCollectionsManagerRequester.EXPECT().
				FindAgentByAddress(gomock.Any(), enums.AddressTypeEvm, gomock.Any()).
				Return(foundCollection, uint64(9), "", nil)
		}

		t.Run("Should fall back to the manager's collection and token", func(t *testing.T) {
			t.Parallel()

			suite := WhenResolvingMintedTokenBeforeEach(t)
			initSuite(suite)

			collection, tokenId, err := suite.sut.ResolveMintedToken(context.Background(), agent)

			assert.NoError(t, err)
			assert.Equal(t, uint64(9), tokenId)
			assert.Equal(t, foundCollection, collection)
		})
	})

	t.Run("Given no token is found anywhere", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenResolvingMintedTokenTestingSuite) {
			suite.mockWalletCredentialsStorage.EXPECT().
				GetForId(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&identity_interfaces.WalletCredentials{Address: nodeWallet}, true, nil)
			suite.mockAgentCollectionRequester.EXPECT().
				AddressToTokenId(gomock.Any(), gomock.Any(), enums.AddressTypeEvm, gomock.Any()).
				Return(uint64(0), nil)
			suite.mockAgentCollectionsManagerRequester.EXPECT().
				FindAgentByAddress(gomock.Any(), enums.AddressTypeEvm, gomock.Any()).
				Return(common.Address{}, uint64(0), "", nil)
		}

		t.Run("Should report tokenId 0", func(t *testing.T) {
			t.Parallel()

			suite := WhenResolvingMintedTokenBeforeEach(t)
			initSuite(suite)

			_, tokenId, err := suite.sut.ResolveMintedToken(context.Background(), agent)

			assert.NoError(t, err)
			assert.Equal(t, uint64(0), tokenId)
		})
	})
}
