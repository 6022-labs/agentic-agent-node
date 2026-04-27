package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/services"
	"github.com/6022-labs/agentic-agent-node/src/common/models"
	common_mocks "github.com/6022-labs/agentic-agent-node/tests/common_mocks/interfaces_mocks"
	chain_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestWhenAuthorizingNftOwner(t *testing.T) {
	t.Parallel()

	collectionAddr := common.HexToAddress("0xCCCCcccccccccccccccccccccccccccccccccccc")
	owner := common.HexToAddress("0x4444444444444444444444444444444444444444")
	identity := &models.AgentIdentity{
		CollectionAddress: collectionAddr.Hex(),
		TokenId:           42,
		Name:              "agent.eth",
	}

	t.Run("Given the identity resolver fails", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		resolver := common_mocks.NewMockAgentIdentityResolverInterface(ctrl)
		resolver.EXPECT().Resolve(gomock.Any()).Return(nil, services.ErrIdentityNotFound)
		collection := chain_mocks.NewMockAgentCollectionRequesterInterface(ctrl)

		authorizer := services.NewNftOwnerAuthorizer(resolver, collection)

		t.Run("Should return false and propagate the error", func(t *testing.T) {
			ok, err := authorizer.IsAuthorized(context.Background(), owner)
			assert.False(t, ok)
			assert.ErrorIs(t, err, services.ErrIdentityNotFound)
		})
	})

	t.Run("Given the requested address is the current NFT owner", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		resolver := common_mocks.NewMockAgentIdentityResolverInterface(ctrl)
		resolver.EXPECT().Resolve(gomock.Any()).Return(identity, nil)
		collection := chain_mocks.NewMockAgentCollectionRequesterInterface(ctrl)
		collection.EXPECT().OwnerOf(collectionAddr, uint64(42)).Return(owner, nil)

		authorizer := services.NewNftOwnerAuthorizer(resolver, collection)

		t.Run("Should authorize", func(t *testing.T) {
			ok, err := authorizer.IsAuthorized(context.Background(), owner)
			assert.NoError(t, err)
			assert.True(t, ok)
		})
	})

	t.Run("Given the requested address is not the NFT owner", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		resolver := common_mocks.NewMockAgentIdentityResolverInterface(ctrl)
		resolver.EXPECT().Resolve(gomock.Any()).Return(identity, nil)
		collection := chain_mocks.NewMockAgentCollectionRequesterInterface(ctrl)
		collection.EXPECT().OwnerOf(collectionAddr, uint64(42)).Return(common.HexToAddress("0xdeadbeef"), nil)

		authorizer := services.NewNftOwnerAuthorizer(resolver, collection)

		t.Run("Should reject", func(t *testing.T) {
			ok, err := authorizer.IsAuthorized(context.Background(), owner)
			assert.NoError(t, err)
			assert.False(t, ok)
		})
	})

	t.Run("Given OwnerOf RPC fails", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		resolver := common_mocks.NewMockAgentIdentityResolverInterface(ctrl)
		resolver.EXPECT().Resolve(gomock.Any()).Return(identity, nil)
		collection := chain_mocks.NewMockAgentCollectionRequesterInterface(ctrl)
		collection.EXPECT().OwnerOf(collectionAddr, uint64(42)).Return(common.Address{}, errors.New("rpc"))

		authorizer := services.NewNftOwnerAuthorizer(resolver, collection)

		t.Run("Should propagate the error", func(t *testing.T) {
			ok, err := authorizer.IsAuthorized(context.Background(), owner)
			assert.False(t, ok)
			assert.ErrorContains(t, err, "rpc")
		})
	})
}
