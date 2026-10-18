package services_test

import (
	"context"
	"errors"
	"testing"

	identity_enums "github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/services"
	repository_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	chain_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenAuthorizingNftOwnerTestingSuite struct {
	sut *services.NftOwnerAuthorizer

	mockAgentRepository                  *repository_mocks.MockAgentRepositoryInterface
	mockAgentRegistrationStateRepository *repository_mocks.MockAgentRegistrationStateRepositoryInterface
	mockAgentCollectionRequester         *chain_mocks.MockAgentCollectionRequesterInterface
}

func WhenAuthorizingNftOwnerBeforeEach(t *testing.T) *WhenAuthorizingNftOwnerTestingSuite {
	mockController := gomock.NewController(t)

	mockAgentRepository := repository_mocks.NewMockAgentRepositoryInterface(mockController)
	mockAgentRegistrationStateRepository := repository_mocks.NewMockAgentRegistrationStateRepositoryInterface(
		mockController,
	)
	mockAgentCollectionRequester := chain_mocks.NewMockAgentCollectionRequesterInterface(mockController)

	sut := services.NewNftOwnerAuthorizer(
		mockAgentRepository,
		mockAgentRegistrationStateRepository,
		mockAgentCollectionRequester,
	)

	return &WhenAuthorizingNftOwnerTestingSuite{
		sut: sut,

		mockAgentRepository:                  mockAgentRepository,
		mockAgentRegistrationStateRepository: mockAgentRegistrationStateRepository,
		mockAgentCollectionRequester:         mockAgentCollectionRequester,
	}
}

func TestWhenAuthorizingNftOwner(t *testing.T) {
	t.Parallel()

	collectionAddr := common.HexToAddress("0xCCCCcccccccccccccccccccccccccccccccccccc")
	owner := common.HexToAddress("0x4444444444444444444444444444444444444444")

	mintedState := func() interface{} {
		return model_builders.NewAgentRegistrationStateBuilder().
			WithMintStatus(identity_enums.TransactionStatusConfirmed).
			Build()
	}

	t.Run("Given the mint is not yet confirmed", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAuthorizingNftOwnerTestingSuite) {
			state := model_builders.NewAgentRegistrationStateBuilder().
				WithMintStatus(identity_enums.TransactionStatusPending).
				Build()
			agent := model_builders.NewAgentBuilder().WithCollectionTokenId(42).Build()
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockAgentRegistrationStateRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(state, nil)
		}

		t.Run("Should authorize the owner", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingNftOwnerBeforeEach(t)
			initSuite(suite)

			ok, err := suite.sut.IsAuthorized(context.Background(), owner)

			assert.NoError(t, err)
			assert.True(t, ok)
		})
	})

	t.Run("Given the registration state lookup fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAuthorizingNftOwnerTestingSuite) {
			agent := model_builders.NewAgentBuilder().WithCollectionTokenId(42).Build()
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockAgentRegistrationStateRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(nil, errors.New("db down"))
		}

		t.Run("Should return false and propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenAuthorizingNftOwnerBeforeEach(t)
			initSuite(suite)

			ok, err := suite.sut.IsAuthorized(context.Background(), owner)

			assert.False(t, ok)
			assert.ErrorContains(t, err, "db down")
		})
	})

	t.Run("Given the owner lookup fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAuthorizingNftOwnerTestingSuite) {
			agent := model_builders.NewAgentBuilder().
				WithCollectionAddress(collectionAddr.Hex()).
				WithCollectionTokenId(42).
				Build()
			suite.mockAgentRegistrationStateRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(mintedState(), nil)
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockAgentCollectionRequester.EXPECT().
				OwnerOf(gomock.Any(), collectionAddr, uint64(42)).
				Return(common.Address{}, errors.New("rpc"))
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

	t.Run("Given the mint is confirmed and the requested address is not the NFT owner", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAuthorizingNftOwnerTestingSuite) {
			agent := model_builders.NewAgentBuilder().
				WithCollectionAddress(collectionAddr.Hex()).
				WithCollectionTokenId(42).
				Build()
			suite.mockAgentRegistrationStateRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(mintedState(), nil)
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockAgentCollectionRequester.EXPECT().
				OwnerOf(gomock.Any(), collectionAddr, uint64(42)).
				Return(common.HexToAddress("0xdeadbeef"), nil)
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

	t.Run("Given the mint is confirmed and the requested address is the NFT owner", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAuthorizingNftOwnerTestingSuite) {
			agent := model_builders.NewAgentBuilder().
				WithCollectionAddress(collectionAddr.Hex()).
				WithCollectionTokenId(42).
				Build()
			suite.mockAgentRegistrationStateRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(mintedState(), nil)
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockAgentCollectionRequester.EXPECT().
				OwnerOf(gomock.Any(), collectionAddr, uint64(42)).
				Return(owner, nil)
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
}
