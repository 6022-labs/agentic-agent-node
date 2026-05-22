package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/services"
	"github.com/6022-labs/agentic-agent-node/src/common/models"
	chain_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	common_mocks "github.com/6022-labs/agentic-agent-node/tests/common_mocks/interfaces_mocks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenAuthorizingNftOwnerTestingSuite struct {
	sut *services.NftOwnerAuthorizer

	mockResolver   *common_mocks.MockAgentIdentityResolverInterface
	mockCollection *chain_mocks.MockAgentCollectionRequesterInterface
}

func WhenAuthorizingNftOwnerBeforeEach(t *testing.T) *WhenAuthorizingNftOwnerTestingSuite {
	mockController := gomock.NewController(t)

	mockResolver := common_mocks.NewMockAgentIdentityResolverInterface(mockController)
	mockCollection := chain_mocks.NewMockAgentCollectionRequesterInterface(mockController)

	sut := services.NewNftOwnerAuthorizer(mockResolver, mockCollection)

	return &WhenAuthorizingNftOwnerTestingSuite{
		sut: sut,

		mockResolver:   mockResolver,
		mockCollection: mockCollection,
	}
}

func TestWhenAuthorizingNftOwner(t *testing.T) {
	t.Parallel()

	collectionAddr := common.HexToAddress("0xCCCCcccccccccccccccccccccccccccccccccccc")
	owner := common.HexToAddress("0x4444444444444444444444444444444444444444")
	identity := &models.AgentIdentity{
		CollectionAddress: collectionAddr.Hex(),
		TokenId:           42,
		Name:              "agent.eth",
	}

	t.Run("Given the agent has no on-chain identity yet", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAuthorizingNftOwnerTestingSuite) {
			suite.mockResolver.EXPECT().Resolve(gomock.Any()).Return(nil, services.ErrIdentityNotFound)
		}

		t.Run("Should authorize while unclaimed", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingNftOwnerBeforeEach(t)
			initSuite(suite)

			ok, err := suite.sut.IsAuthorized(context.Background(), owner)

			assert.NoError(t, err)
			assert.True(t, ok)
		})
	})

	t.Run("Given the identity resolver fails for another reason", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAuthorizingNftOwnerTestingSuite) {
			suite.mockResolver.EXPECT().Resolve(gomock.Any()).Return(nil, errors.New("wallet not loaded yet"))
		}

		t.Run("Should return false and propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingNftOwnerBeforeEach(t)
			initSuite(suite)

			ok, err := suite.sut.IsAuthorized(context.Background(), owner)

			assert.False(t, ok)
			assert.ErrorContains(t, err, "wallet not loaded yet")
		})
	})

	t.Run("Given the requested address is the current NFT owner", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAuthorizingNftOwnerTestingSuite) {
			suite.mockResolver.EXPECT().Resolve(gomock.Any()).Return(identity, nil)
			suite.mockCollection.EXPECT().OwnerOf(collectionAddr, uint64(42)).Return(owner, nil)
		}

		t.Run("Should authorize", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingNftOwnerBeforeEach(t)
			initSuite(suite)

			ok, err := suite.sut.IsAuthorized(context.Background(), owner)

			assert.NoError(t, err)
			assert.True(t, ok)
		})
	})

	t.Run("Given the requested address is not the NFT owner", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAuthorizingNftOwnerTestingSuite) {
			suite.mockResolver.EXPECT().Resolve(gomock.Any()).Return(identity, nil)
			suite.mockCollection.EXPECT().OwnerOf(collectionAddr, uint64(42)).Return(common.HexToAddress("0xdeadbeef"), nil)
		}

		t.Run("Should reject", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingNftOwnerBeforeEach(t)
			initSuite(suite)

			ok, err := suite.sut.IsAuthorized(context.Background(), owner)

			assert.NoError(t, err)
			assert.False(t, ok)
		})
	})

	t.Run("Given OwnerOf RPC fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAuthorizingNftOwnerTestingSuite) {
			suite.mockResolver.EXPECT().Resolve(gomock.Any()).Return(identity, nil)
			suite.mockCollection.EXPECT().OwnerOf(collectionAddr, uint64(42)).Return(common.Address{}, errors.New("rpc"))
		}

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingNftOwnerBeforeEach(t)
			initSuite(suite)

			ok, err := suite.sut.IsAuthorized(context.Background(), owner)

			assert.False(t, ok)
			assert.ErrorContains(t, err, "rpc")
		})
	})
}
