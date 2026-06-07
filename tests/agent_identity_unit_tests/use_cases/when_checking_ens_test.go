package use_cases_test

import (
	"context"
	"testing"

	identity_enums "github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	repository_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	interface_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenCheckingEnsTestingSuite struct {
	sut *use_cases.CheckAgentEns

	mockEnsPublisher                     *interface_mocks.MockEnsPublisherInterface
	mockAgentRepository                  *repository_mocks.MockAgentRepositoryInterface
	mockAgentRegistrationStateRepository *repository_mocks.MockAgentRegistrationStateRepositoryInterface
	mockRegistrationReconciler           *interface_mocks.MockRegistrationReconcilerInterface
}

func WhenCheckingEnsBeforeEach(t *testing.T) *WhenCheckingEnsTestingSuite {
	mockController := gomock.NewController(t)

	mockEnsPublisher := interface_mocks.NewMockEnsPublisherInterface(mockController)
	mockAgentRepository := repository_mocks.NewMockAgentRepositoryInterface(mockController)
	mockAgentRegistrationStateRepository := repository_mocks.NewMockAgentRegistrationStateRepositoryInterface(mockController)
	mockRegistrationReconciler := interface_mocks.NewMockRegistrationReconcilerInterface(mockController)

	sut := use_cases.NewCheckAgentEns(
		mockEnsPublisher,
		mockAgentRepository,
		mockAgentRegistrationStateRepository,
		mockRegistrationReconciler,
	)

	return &WhenCheckingEnsTestingSuite{
		sut:                                  sut,
		mockEnsPublisher:                     mockEnsPublisher,
		mockAgentRepository:                  mockAgentRepository,
		mockAgentRegistrationStateRepository: mockAgentRegistrationStateRepository,
		mockRegistrationReconciler:           mockRegistrationReconciler,
	}
}

func TestWhenCheckingEns(t *testing.T) {
	t.Parallel()

	t.Run("Given the agent is not minted yet", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCheckingEnsTestingSuite) {
			agent := model_builders.NewAgentBuilder().WithoutCollectionTokenId().Build()
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
		}

		t.Run("Should return a Conflict without reading the chain", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingEnsBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background())

			var conflictErr *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictErr)
		})
	})

	t.Run("Given ENS is already confirmed in the database", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCheckingEnsTestingSuite) {
			agent := model_builders.NewAgentBuilder().WithCollectionTokenId(7).Build()
			confirmedState := model_builders.NewAgentRegistrationStateBuilder().WithEnsRegistryStatus(identity_enums.TransactionStatusConfirmed).Build()

			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockAgentRegistrationStateRepository.EXPECT().Get(gomock.Any()).Return(confirmedState, nil)
		}

		t.Run("Should return confirmed without reading the chain", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingEnsBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, string(identity_enums.TransactionStatusConfirmed), resp.EnsStatus)
		})
	})

	t.Run("Given a provisioned subnode not yet marked confirmed", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCheckingEnsTestingSuite) {
			agent := model_builders.NewAgentBuilder().WithCollectionTokenId(7).Build()
			pendingState := model_builders.NewAgentRegistrationStateBuilder().WithEnsRegistryStatus(identity_enums.TransactionStatusPending).Build()
			confirmedState := model_builders.NewAgentRegistrationStateBuilder().WithEnsRegistryStatus(identity_enums.TransactionStatusConfirmed).Build()

			suite.mockEnsPublisher.EXPECT().IsProvisioned(gomock.Any()).Return(true, nil)
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockAgentRegistrationStateRepository.EXPECT().Get(gomock.Any()).Return(pendingState, nil)
			suite.mockRegistrationReconciler.EXPECT().ReconcileEnsRegistry(gomock.Any(), gomock.Any()).Return(confirmedState, nil)
		}

		t.Run("Should reconcile ENS and report confirmed", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingEnsBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, string(identity_enums.TransactionStatusConfirmed), resp.EnsStatus)
		})
	})

	t.Run("Given the subnode is not provisioned yet", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCheckingEnsTestingSuite) {
			agent := model_builders.NewAgentBuilder().WithCollectionTokenId(7).Build()
			pendingState := model_builders.NewAgentRegistrationStateBuilder().WithEnsRegistryStatus(identity_enums.TransactionStatusPending).Build()

			suite.mockEnsPublisher.EXPECT().IsProvisioned(gomock.Any()).Return(false, nil)
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockAgentRegistrationStateRepository.EXPECT().Get(gomock.Any()).Return(pendingState, nil)
		}

		t.Run("Should report ENS still pending without reconciling", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingEnsBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, string(identity_enums.TransactionStatusPending), resp.EnsStatus)
		})
	})
}
