package services_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	identity_enums "github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/services"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	repository_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	interface_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/ethereum/go-ethereum/common"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenReconcilingIdentityTestingSuite struct {
	sut *services.RegistrationReconciler

	mockAgentCollectionRequester         *interface_mocks.MockAgentCollectionRequesterInterface
	mockAgentRepository                  *repository_mocks.MockAgentRepositoryInterface
	mockAgentRegistrationStateRepository *repository_mocks.MockAgentRegistrationStateRepositoryInterface
}

func WhenReconcilingIdentityBeforeEach(t *testing.T) *WhenReconcilingIdentityTestingSuite {
	mockController := gomock.NewController(t)

	mockAgentCollectionRequester := interface_mocks.NewMockAgentCollectionRequesterInterface(mockController)
	mockAgentRepository := repository_mocks.NewMockAgentRepositoryInterface(mockController)
	mockAgentRegistrationStateRepository := repository_mocks.NewMockAgentRegistrationStateRepositoryInterface(
		mockController,
	)

	sut := services.NewRegistrationReconciler(
		mockAgentRepository,
		mockAgentCollectionRequester,
		mockAgentRegistrationStateRepository,
	)

	return &WhenReconcilingIdentityTestingSuite{
		sut:                                  sut,
		mockAgentCollectionRequester:         mockAgentCollectionRequester,
		mockAgentRepository:                  mockAgentRepository,
		mockAgentRegistrationStateRepository: mockAgentRegistrationStateRepository,
	}
}

func TestWhenReconcilingIdentity(t *testing.T) {
	t.Parallel()

	collection := common_tests.RandomEthAddress()

	mintedAgent := func() *models.Agent {
		return model_builders.NewAgentBuilder().
			WithChainId("80002").
			WithCollectionAddress(collection.Hex()).
			WithCollectionTokenId(5).
			Build()
	}
	mintedState := func() *models.AgentRegistrationState {
		return model_builders.NewAgentRegistrationStateBuilder().
			WithMintStatus(identity_enums.TransactionStatusConfirmed).
			Build()
	}

	t.Run("Given the agent lookup fails", func(t *testing.T) {
		t.Parallel()

		lookupErr := errors.New("db down")

		initSuite := func(suite *WhenReconcilingIdentityTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(nil, lookupErr)
		}

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenReconcilingIdentityBeforeEach(t)
			initSuite(suite)

			err := suite.sut.ReconcileIdentity(context.Background())

			assert.ErrorIs(t, err, lookupErr)
		})
	})

	t.Run("Given no agent row exists", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenReconcilingIdentityTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(nil, nil)
		}

		t.Run("Should return nil without reading the chain", func(t *testing.T) {
			t.Parallel()

			suite := WhenReconcilingIdentityBeforeEach(t)
			initSuite(suite)

			err := suite.sut.ReconcileIdentity(context.Background())

			assert.NoError(t, err)
		})
	})

	t.Run("Given the agent is not minted yet", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenReconcilingIdentityTestingSuite) {
			suite.mockAgentRepository.EXPECT().
				Get(gomock.Any()).
				Return(model_builders.NewAgentBuilder().WithoutCollectionTokenId().Build(), nil)
		}

		t.Run("Should return nil without reading the chain", func(t *testing.T) {
			t.Parallel()

			suite := WhenReconcilingIdentityBeforeEach(t)
			initSuite(suite)

			err := suite.sut.ReconcileIdentity(context.Background())

			assert.NoError(t, err)
		})
	})

	t.Run("Given the registration state lookup fails", func(t *testing.T) {
		t.Parallel()

		lookupErr := errors.New("db down")

		initSuite := func(suite *WhenReconcilingIdentityTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(mintedAgent(), nil)
			suite.mockAgentRegistrationStateRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(nil, lookupErr)
		}

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenReconcilingIdentityBeforeEach(t)
			initSuite(suite)

			err := suite.sut.ReconcileIdentity(context.Background())

			assert.ErrorIs(t, err, lookupErr)
		})
	})

	t.Run("Given no registration state exists", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenReconcilingIdentityTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(mintedAgent(), nil)
			suite.mockAgentRegistrationStateRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(nil, nil)
		}

		t.Run("Should return nil without reading the chain", func(t *testing.T) {
			t.Parallel()

			suite := WhenReconcilingIdentityBeforeEach(t)
			initSuite(suite)

			err := suite.sut.ReconcileIdentity(context.Background())

			assert.NoError(t, err)
		})
	})

	t.Run("Given the mint is not yet confirmed", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenReconcilingIdentityTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(mintedAgent(), nil)
			suite.mockAgentRegistrationStateRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(model_builders.NewAgentRegistrationStateBuilder().
					WithMintStatus(identity_enums.TransactionStatusPending).
					Build(), nil)
		}

		t.Run("Should return nil without reading the chain", func(t *testing.T) {
			t.Parallel()

			suite := WhenReconcilingIdentityBeforeEach(t)
			initSuite(suite)

			err := suite.sut.ReconcileIdentity(context.Background())

			assert.NoError(t, err)
		})
	})

	t.Run("Given the chain read fails", func(t *testing.T) {
		t.Parallel()

		rpcErr := errors.New("rpc")

		initSuite := func(suite *WhenReconcilingIdentityTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(mintedAgent(), nil)
			suite.mockAgentRegistrationStateRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(mintedState(), nil)
			suite.mockAgentCollectionRequester.EXPECT().
				InformationOf("80002", collection, uint64(5)).
				Return(identity_interfaces.OnChainAgentInfo{}, rpcErr)
		}

		t.Run("Should propagate the error without writing the row", func(t *testing.T) {
			t.Parallel()

			suite := WhenReconcilingIdentityBeforeEach(t)
			initSuite(suite)

			err := suite.sut.ReconcileIdentity(context.Background())

			assert.ErrorIs(t, err, rpcErr)
		})
	})

	t.Run("Given the owner lookup fails", func(t *testing.T) {
		t.Parallel()

		rpcErr := errors.New("rpc")

		initSuite := func(suite *WhenReconcilingIdentityTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(mintedAgent(), nil)
			suite.mockAgentRegistrationStateRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(mintedState(), nil)
			suite.mockAgentCollectionRequester.EXPECT().
				InformationOf("80002", collection, uint64(5)).
				Return(identity_interfaces.OnChainAgentInfo{}, nil)
			suite.mockAgentCollectionRequester.EXPECT().
				OwnerOf("80002", collection, uint64(5)).
				Return(common.Address{}, rpcErr)
		}

		t.Run("Should propagate the error without writing the row", func(t *testing.T) {
			t.Parallel()

			suite := WhenReconcilingIdentityBeforeEach(t)
			initSuite(suite)

			err := suite.sut.ReconcileIdentity(context.Background())

			assert.ErrorIs(t, err, rpcErr)
		})
	})

	t.Run("Given the row update fails", func(t *testing.T) {
		t.Parallel()

		updateErr := errors.New("db down")

		initSuite := func(suite *WhenReconcilingIdentityTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(mintedAgent(), nil)
			suite.mockAgentRegistrationStateRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(mintedState(), nil)
			suite.mockAgentCollectionRequester.EXPECT().
				InformationOf("80002", collection, uint64(5)).
				Return(identity_interfaces.OnChainAgentInfo{}, nil)
			suite.mockAgentCollectionRequester.EXPECT().
				OwnerOf("80002", collection, uint64(5)).
				Return(common_tests.RandomEthAddress(), nil)
			suite.mockAgentRepository.EXPECT().UpdateOnChainIdentity(gomock.Any(), gomock.Any()).Return(updateErr)
		}

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenReconcilingIdentityBeforeEach(t)
			initSuite(suite)

			err := suite.sut.ReconcileIdentity(context.Background())

			assert.ErrorIs(t, err, updateErr)
		})
	})
	t.Run("Given a minted agent", func(t *testing.T) {
		t.Parallel()

		creator := common_tests.RandomEthAddress()
		nftOwner := common_tests.RandomEthAddress()
		info := identity_interfaces.OnChainAgentInfo{
			Name:       "Alice",
			Creator:    creator,
			CloneOf:    lo.ToPtr(uint64(3)),
			Images:     map[string]string{"avatar": "ipfs://new"},
			Attributes: map[string]string{"role": "expert"},
		}

		initSuite := func(suite *WhenReconcilingIdentityTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(mintedAgent(), nil)
			suite.mockAgentRegistrationStateRepository.EXPECT().
				GetByAgentId(gomock.Any(), gomock.Any()).
				Return(mintedState(), nil)
			suite.mockAgentCollectionRequester.EXPECT().
				InformationOf("80002", collection, uint64(5)).
				Return(info, nil)
			suite.mockAgentCollectionRequester.EXPECT().
				OwnerOf("80002", collection, uint64(5)).
				Return(nftOwner, nil)
		}

		t.Run("Should write the on-chain owner and NFT data", func(t *testing.T) {
			t.Parallel()

			suite := WhenReconcilingIdentityBeforeEach(t)
			initSuite(suite)
			suite.mockAgentRepository.EXPECT().
				UpdateOnChainIdentity(gomock.Any(), gomock.Cond(func(a *models.Agent) bool {
					return lo.FromPtr(a.Owner) == nftOwner.Hex() &&
						lo.FromPtr(a.Creator) == creator.Hex() &&
						lo.FromPtr(a.Name) == "Alice" &&
						lo.FromPtr(a.CloneOf) == 3 &&
						a.Images["avatar"] == "ipfs://new" &&
						a.Attributes["role"] == "expert"
				})).
				Return(nil)

			err := suite.sut.ReconcileIdentity(context.Background())

			assert.NoError(t, err)
		})

		t.Run("Should recompose the soul from the on-chain role", func(t *testing.T) {
			t.Parallel()

			suite := WhenReconcilingIdentityBeforeEach(t)
			initSuite(suite)
			suite.mockAgentRepository.EXPECT().
				UpdateOnChainIdentity(gomock.Any(), gomock.Cond(func(a *models.Agent) bool {
					soul := lo.FromPtr(a.Soul)
					return strings.Contains(soul, "I am Alice") && strings.Contains(soul, `role "expert"`)
				})).
				Return(nil)

			err := suite.sut.ReconcileIdentity(context.Background())

			assert.NoError(t, err)
		})
	})

}
