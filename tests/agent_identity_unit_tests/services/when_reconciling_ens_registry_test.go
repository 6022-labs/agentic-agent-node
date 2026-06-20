package services_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	identity_enums "github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/services"
	repository_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	interface_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenReconcilingEnsRegistryTestingSuite struct {
	sut *services.RegistrationReconciler

	mockAgentCollectionRequester         *interface_mocks.MockAgentCollectionRequesterInterface
	mockAgentRepository                  *repository_mocks.MockAgentRepositoryInterface
	mockAgentRegistrationStateRepository *repository_mocks.MockAgentRegistrationStateRepositoryInterface
}

func WhenReconcilingEnsRegistryBeforeEach(t *testing.T) *WhenReconcilingEnsRegistryTestingSuite {
	mockController := gomock.NewController(t)

	mockAgentCollectionRequester := interface_mocks.NewMockAgentCollectionRequesterInterface(mockController)
	mockAgentRepository := repository_mocks.NewMockAgentRepositoryInterface(mockController)
	mockAgentRegistrationStateRepository := repository_mocks.NewMockAgentRegistrationStateRepositoryInterface(mockController)

	sut := services.NewRegistrationReconciler(mockAgentRepository, mockAgentCollectionRequester, mockAgentRegistrationStateRepository)

	return &WhenReconcilingEnsRegistryTestingSuite{
		sut:                                  sut,
		mockAgentCollectionRequester:         mockAgentCollectionRequester,
		mockAgentRepository:                  mockAgentRepository,
		mockAgentRegistrationStateRepository: mockAgentRegistrationStateRepository,
	}
}

func TestWhenReconcilingEnsRegistry(t *testing.T) {
	t.Parallel()

	t.Run("Given a provisioned ENS subnode", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenReconcilingEnsRegistryTestingSuite) {
			suite.mockAgentRegistrationStateRepository.EXPECT().Upsert(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, s *models.AgentRegistrationState) (*models.AgentRegistrationState, error) {
					return s, nil
				})
		}

		t.Run("Should flip the ENS status to confirmed and stamp the time", func(t *testing.T) {
			t.Parallel()

			suite := WhenReconcilingEnsRegistryBeforeEach(t)
			initSuite(suite)

			state := model_builders.NewAgentRegistrationStateBuilder().
				WithEnsRegistryStatus(identity_enums.TransactionStatusPending).
				Build()
			savedState, err := suite.sut.ReconcileEnsRegistry(context.Background(), uuid.New(), state)

			assert.NoError(t, err)
			assert.Equal(t, identity_enums.TransactionStatusConfirmed, savedState.EnsRegistryStatus)
			assert.NotNil(t, savedState.EnsPublishedAt)
		})
	})
}
