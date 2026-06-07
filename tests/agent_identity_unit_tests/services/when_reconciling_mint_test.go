package services_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	identity_enums "github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/services"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	repository_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	interface_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenReconcilingMintTestingSuite struct {
	sut *services.RegistrationReconciler

	mockCollection      *interface_mocks.MockAgentCollectionRequesterInterface
	mockAgentRepository *repository_mocks.MockAgentRepositoryInterface
	mockStateRepository *repository_mocks.MockAgentRegistrationStateRepositoryInterface
}

func WhenReconcilingMintBeforeEach(t *testing.T) *WhenReconcilingMintTestingSuite {
	mockController := gomock.NewController(t)

	mockCollection := interface_mocks.NewMockAgentCollectionRequesterInterface(mockController)
	mockAgentRepository := repository_mocks.NewMockAgentRepositoryInterface(mockController)
	mockStateRepository := repository_mocks.NewMockAgentRegistrationStateRepositoryInterface(mockController)

	sut := services.NewRegistrationReconciler(mockAgentRepository, mockCollection, mockStateRepository)

	return &WhenReconcilingMintTestingSuite{
		sut:                 sut,
		mockCollection:      mockCollection,
		mockAgentRepository: mockAgentRepository,
		mockStateRepository: mockStateRepository,
	}
}

func TestWhenReconcilingMint(t *testing.T) {
	t.Parallel()

	t.Run("Given a freshly minted token", func(t *testing.T) {
		t.Parallel()

		creator := common_tests.RandomEthAddress()
		nftOwner := common_tests.RandomEthAddress()

		initSuite := func(suite *WhenReconcilingMintTestingSuite) {
			suite.mockCollection.EXPECT().InformationOf(gomock.Any(), uint64(5)).Return(identity_interfaces.OnChainAgentInfo{Creator: creator}, nil)
			suite.mockCollection.EXPECT().OwnerOf(gomock.Any(), uint64(5)).Return(nftOwner, nil)
			suite.mockAgentRepository.EXPECT().Upsert(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, a *models.Agent) (*models.Agent, error) { return a, nil })
			suite.mockStateRepository.EXPECT().Upsert(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, s *models.AgentRegistrationState) (*models.AgentRegistrationState, error) {
					return s, nil
				})
		}

		t.Run("Should write the tokenId + confirmed mint status onto the rows", func(t *testing.T) {
			t.Parallel()

			suite := WhenReconcilingMintBeforeEach(t)
			initSuite(suite)

			agent := model_builders.NewAgentBuilder().WithoutCollectionTokenId().Build()
			savedAgent, savedState, err := suite.sut.ReconcileMint(context.Background(), agent, nil, common_tests.RandomEthAddress(), 5, nil)

			assert.NoError(t, err)
			assert.NotNil(t, savedAgent.CollectionTokenId)
			assert.Equal(t, uint64(5), *savedAgent.CollectionTokenId)
			assert.Equal(t, creator.Hex(), savedAgent.Creator)
			assert.Equal(t, identity_enums.TransactionStatusConfirmed, savedState.MintStatus)
			assert.NotNil(t, savedState.MintedAt)
		})
	})
}
