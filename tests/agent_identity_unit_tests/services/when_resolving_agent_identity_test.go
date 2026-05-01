package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/services"
	"github.com/6022-labs/agentic-agent-node/src/common/models"
	"github.com/6022-labs/agentic-agent-node/src/common/models/enums"
	identity_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type WhenResolvingAgentIdentityTestingSuite struct {
	sut *services.AgentIdentityResolver

	mockState      *identity_mocks.MockBootstrapStateHolderInterface
	mockCollection *identity_mocks.MockAgentCollectionRequesterInterface
	mockManager    *identity_mocks.MockAgentCollectionsManagerRequesterInterface
}

func WhenResolvingAgentIdentityBeforeEach(t *testing.T) *WhenResolvingAgentIdentityTestingSuite {
	mockController := gomock.NewController(t)

	mockState := identity_mocks.NewMockBootstrapStateHolderInterface(mockController)
	mockCollection := identity_mocks.NewMockAgentCollectionRequesterInterface(mockController)
	mockManager := identity_mocks.NewMockAgentCollectionsManagerRequesterInterface(mockController)

	sut := services.NewAgentIdentityResolver(zap.NewNop(), mockState, mockCollection, mockManager)

	return &WhenResolvingAgentIdentityTestingSuite{
		sut: sut,

		mockState:      mockState,
		mockCollection: mockCollection,
		mockManager:    mockManager,
	}
}

func TestWhenResolvingAgentIdentity(t *testing.T) {
	t.Parallel()

	walletAddr := common.HexToAddress("0x5555555555555555555555555555555555555555")
	collectionA := common.HexToAddress("0xAAAAAAaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	collectionB := common.HexToAddress("0xBBBBBBbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")

	t.Run("Given the wallet has not been loaded yet", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenResolvingAgentIdentityTestingSuite) {
			suite.mockState.EXPECT().Get().Return(models.BootstrapState{Status: enums.BootstrapStatusAwaitingVault})
		}

		t.Run("Should return an error and not call any chain methods", func(t *testing.T) {
			t.Parallel()

			suite := WhenResolvingAgentIdentityBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Resolve(context.Background())

			assert.Error(t, err)
		})
	})

	t.Run("Given no collection holds a token for the wallet", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenResolvingAgentIdentityTestingSuite) {
			suite.mockState.EXPECT().Get().Return(models.BootstrapState{AgentAddress: walletAddr.Hex()})
			suite.mockManager.EXPECT().ListAllCollections().Return([]common.Address{collectionA, collectionB}, nil)
			suite.mockCollection.EXPECT().TokenIdByWallet(collectionA, walletAddr).Return(uint64(0), nil)
			suite.mockCollection.EXPECT().TokenIdByWallet(collectionB, walletAddr).Return(uint64(0), nil)
		}

		t.Run("Should return ErrIdentityNotFound", func(t *testing.T) {
			t.Parallel()

			suite := WhenResolvingAgentIdentityBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Resolve(context.Background())

			assert.ErrorIs(t, err, services.ErrIdentityNotFound)
		})
	})

	t.Run("Given a token is found in the second collection", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenResolvingAgentIdentityTestingSuite) {
			suite.mockState.EXPECT().Get().Return(models.BootstrapState{AgentAddress: walletAddr.Hex()})
			suite.mockManager.EXPECT().ListAllCollections().Return([]common.Address{collectionA, collectionB}, nil)
			suite.mockCollection.EXPECT().TokenIdByWallet(collectionA, walletAddr).Return(uint64(0), nil)
			suite.mockCollection.EXPECT().TokenIdByWallet(collectionB, walletAddr).Return(uint64(7), nil)
			suite.mockCollection.EXPECT().NameOf(collectionB, uint64(7)).Return("agent.eth", nil)
		}

		t.Run("Should resolve to that collection + tokenId + name and serve subsequent calls from cache", func(t *testing.T) {
			t.Parallel()

			suite := WhenResolvingAgentIdentityBeforeEach(t)
			initSuite(suite)

			identity, err := suite.sut.Resolve(context.Background())
			assert.NoError(t, err)
			assert.Equal(t, collectionB.Hex(), identity.CollectionAddress)
			assert.Equal(t, uint64(7), identity.TokenId)
			assert.Equal(t, "agent.eth", identity.Name)

			// second call must hit the cache — the gomock controller would fail
			// if any chain method ran a second time.
			cached, err := suite.sut.Resolve(context.Background())
			assert.NoError(t, err)
			assert.Equal(t, "agent.eth", cached.Name)
		})
	})

	t.Run("Given TokenIdByWallet errors on the first collection", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenResolvingAgentIdentityTestingSuite) {
			suite.mockState.EXPECT().Get().Return(models.BootstrapState{AgentAddress: walletAddr.Hex()})
			suite.mockManager.EXPECT().ListAllCollections().Return([]common.Address{collectionA, collectionB}, nil)
			suite.mockCollection.EXPECT().TokenIdByWallet(collectionA, walletAddr).Return(uint64(0), errors.New("rpc"))
			suite.mockCollection.EXPECT().TokenIdByWallet(collectionB, walletAddr).Return(uint64(3), nil)
			suite.mockCollection.EXPECT().NameOf(collectionB, uint64(3)).Return("ok.eth", nil)
		}

		t.Run("Should skip the failing collection and resolve the next", func(t *testing.T) {
			t.Parallel()

			suite := WhenResolvingAgentIdentityBeforeEach(t)
			initSuite(suite)

			identity, err := suite.sut.Resolve(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, "ok.eth", identity.Name)
		})
	})

	t.Run("Given the manager fails to list collections", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenResolvingAgentIdentityTestingSuite) {
			suite.mockState.EXPECT().Get().Return(models.BootstrapState{AgentAddress: walletAddr.Hex()})
			suite.mockManager.EXPECT().ListAllCollections().Return(nil, errors.New("manager rpc"))
		}

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenResolvingAgentIdentityBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Resolve(context.Background())

			assert.ErrorContains(t, err, "manager rpc")
		})
	})
}
