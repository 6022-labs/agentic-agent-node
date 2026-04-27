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

func TestWhenResolvingAgentIdentity(t *testing.T) {
	t.Parallel()

	walletAddr := common.HexToAddress("0x5555555555555555555555555555555555555555")
	collectionA := common.HexToAddress("0xAAAAAAaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	collectionB := common.HexToAddress("0xBBBBBBbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")

	t.Run("Given the wallet has not been loaded yet", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		state := identity_mocks.NewMockBootstrapStateHolderInterface(ctrl)
		state.EXPECT().Get().Return(models.BootstrapState{Status: enums.BootstrapStatusAwaitingVault})

		resolver := services.NewAgentIdentityResolver(zap.NewNop(), state, nil, nil)

		t.Run("Should return an error and not call any chain methods", func(t *testing.T) {
			_, err := resolver.Resolve(context.Background())
			assert.Error(t, err)
		})
	})

	t.Run("Given no collection holds a token for the wallet", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		state := identity_mocks.NewMockBootstrapStateHolderInterface(ctrl)
		state.EXPECT().Get().Return(models.BootstrapState{AgentAddress: walletAddr.Hex()})
		manager := identity_mocks.NewMockAgentCollectionsManagerRequesterInterface(ctrl)
		manager.EXPECT().ListAllCollections().Return([]common.Address{collectionA, collectionB}, nil)
		collection := identity_mocks.NewMockAgentCollectionRequesterInterface(ctrl)
		collection.EXPECT().TokenIdByWallet(collectionA, walletAddr).Return(uint64(0), nil)
		collection.EXPECT().TokenIdByWallet(collectionB, walletAddr).Return(uint64(0), nil)

		resolver := services.NewAgentIdentityResolver(zap.NewNop(), state, collection, manager)

		t.Run("Should return ErrIdentityNotFound", func(t *testing.T) {
			_, err := resolver.Resolve(context.Background())
			assert.ErrorIs(t, err, services.ErrIdentityNotFound)
		})
	})

	t.Run("Given a token is found in the second collection", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		state := identity_mocks.NewMockBootstrapStateHolderInterface(ctrl)
		state.EXPECT().Get().Return(models.BootstrapState{AgentAddress: walletAddr.Hex()})
		manager := identity_mocks.NewMockAgentCollectionsManagerRequesterInterface(ctrl)
		manager.EXPECT().ListAllCollections().Return([]common.Address{collectionA, collectionB}, nil)
		collection := identity_mocks.NewMockAgentCollectionRequesterInterface(ctrl)
		collection.EXPECT().TokenIdByWallet(collectionA, walletAddr).Return(uint64(0), nil)
		collection.EXPECT().TokenIdByWallet(collectionB, walletAddr).Return(uint64(7), nil)
		collection.EXPECT().NameOf(collectionB, uint64(7)).Return("agent.eth", nil)

		resolver := services.NewAgentIdentityResolver(zap.NewNop(), state, collection, manager)

		t.Run("Should resolve to that collection + tokenId + name", func(t *testing.T) {
			identity, err := resolver.Resolve(context.Background())
			assert.NoError(t, err)
			assert.Equal(t, collectionB.Hex(), identity.CollectionAddress)
			assert.Equal(t, uint64(7), identity.TokenId)
			assert.Equal(t, "agent.eth", identity.Name)
		})

		t.Run("Should serve the second call from cache", func(t *testing.T) {
			// no further EXPECT calls — gomock would fail if any chain method ran.
			identity, err := resolver.Resolve(context.Background())
			assert.NoError(t, err)
			assert.Equal(t, "agent.eth", identity.Name)
		})
	})

	t.Run("Given TokenIdByWallet errors on the first collection", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		state := identity_mocks.NewMockBootstrapStateHolderInterface(ctrl)
		state.EXPECT().Get().Return(models.BootstrapState{AgentAddress: walletAddr.Hex()})
		manager := identity_mocks.NewMockAgentCollectionsManagerRequesterInterface(ctrl)
		manager.EXPECT().ListAllCollections().Return([]common.Address{collectionA, collectionB}, nil)
		collection := identity_mocks.NewMockAgentCollectionRequesterInterface(ctrl)
		collection.EXPECT().TokenIdByWallet(collectionA, walletAddr).Return(uint64(0), errors.New("rpc"))
		collection.EXPECT().TokenIdByWallet(collectionB, walletAddr).Return(uint64(3), nil)
		collection.EXPECT().NameOf(collectionB, uint64(3)).Return("ok.eth", nil)

		resolver := services.NewAgentIdentityResolver(zap.NewNop(), state, collection, manager)

		t.Run("Should skip the failing collection and resolve the next", func(t *testing.T) {
			identity, err := resolver.Resolve(context.Background())
			assert.NoError(t, err)
			assert.Equal(t, "ok.eth", identity.Name)
		})
	})

	t.Run("Given the manager fails to list collections", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		state := identity_mocks.NewMockBootstrapStateHolderInterface(ctrl)
		state.EXPECT().Get().Return(models.BootstrapState{AgentAddress: walletAddr.Hex()})
		manager := identity_mocks.NewMockAgentCollectionsManagerRequesterInterface(ctrl)
		manager.EXPECT().ListAllCollections().Return(nil, errors.New("manager rpc"))

		resolver := services.NewAgentIdentityResolver(zap.NewNop(), state, nil, manager)

		t.Run("Should propagate the error", func(t *testing.T) {
			_, err := resolver.Resolve(context.Background())
			assert.ErrorContains(t, err, "manager rpc")
		})
	})
}
