package use_cases_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	identity_enums "github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	repository_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	interface_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenSelfRegisteringAgentTestingSuite struct {
	sut *use_cases.SelfRegisterAgent

	mockAgentRepository                  *repository_mocks.MockAgentRepositoryInterface
	mockAgentRegistrationStateRepository *repository_mocks.MockAgentRegistrationStateRepositoryInterface
	mockMintPublisher                    *interface_mocks.MockMintPublisherInterface
	mockEnsPublisher                     *interface_mocks.MockEnsPublisherInterface
	mockRegistrationReconciler           *interface_mocks.MockRegistrationReconcilerInterface
}

func WhenSelfRegisteringAgentBeforeEach(t *testing.T) *WhenSelfRegisteringAgentTestingSuite {
	mockController := gomock.NewController(t)

	mockAgentRepository := repository_mocks.NewMockAgentRepositoryInterface(mockController)
	mockAgentRegistrationStateRepository := repository_mocks.NewMockAgentRegistrationStateRepositoryInterface(mockController)
	mockMintPublisher := interface_mocks.NewMockMintPublisherInterface(mockController)
	mockEnsPublisher := interface_mocks.NewMockEnsPublisherInterface(mockController)
	mockRegistrationReconciler := interface_mocks.NewMockRegistrationReconcilerInterface(mockController)

	sut := use_cases.NewSelfRegisterAgent(
		mockAgentRepository,
		mockAgentRegistrationStateRepository,
		mockMintPublisher,
		mockEnsPublisher,
		mockRegistrationReconciler,
	)

	return &WhenSelfRegisteringAgentTestingSuite{
		sut:                                  sut,
		mockAgentRepository:                  mockAgentRepository,
		mockAgentRegistrationStateRepository: mockAgentRegistrationStateRepository,
		mockMintPublisher:                    mockMintPublisher,
		mockEnsPublisher:                     mockEnsPublisher,
		mockRegistrationReconciler:           mockRegistrationReconciler,
	}
}

func TestWhenSelfRegisteringAgent(t *testing.T) {
	t.Parallel()

	owner := common_tests.RandomEthAddress().Hex()

	draft := func() *models.Agent {
		return model_builders.NewAgentBuilder().WithoutCollectionTokenId().WithOwner(owner).Build()
	}

	t.Run("Given the mint lands directly", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSelfRegisteringAgentTestingSuite) {
			mintedAgent := model_builders.NewAgentBuilder().WithCollectionTokenId(5).WithOwner(owner).Build()
			mintedState := model_builders.NewAgentRegistrationStateBuilder().
				WithMintStatus(identity_enums.TransactionStatusConfirmed).
				WithEnsRegistryStatus(identity_enums.TransactionStatusPending).
				Build()
			registeredState := model_builders.NewAgentRegistrationStateBuilder().
				WithMintStatus(identity_enums.TransactionStatusConfirmed).
				WithEnsRegistryStatus(identity_enums.TransactionStatusConfirmed).
				Build()

			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(draft(), nil)
			suite.mockAgentRegistrationStateRepository.EXPECT().Get(gomock.Any()).Return(nil, nil)
			suite.mockMintPublisher.EXPECT().SelfMint(gomock.Any(), gomock.Any(), gomock.Any()).Return(
				&identity_interfaces.MintResult{Agent: mintedAgent, State: mintedState}, nil)
			suite.mockEnsPublisher.EXPECT().PublishRuntimeRecords(gomock.Any()).Return(nil)
			suite.mockRegistrationReconciler.EXPECT().ReconcileEnsRegistry(gomock.Any(), gomock.Any()).Return(registeredState, nil)
		}

		t.Run("Should confirm both mint and ENS", func(t *testing.T) {
			t.Parallel()

			suite := WhenSelfRegisteringAgentBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, string(identity_enums.TransactionStatusConfirmed), resp.MintStatus)
			assert.Equal(t, string(identity_enums.TransactionStatusConfirmed), resp.EnsStatus)
		})
	})

	t.Run("Given the mint is left pending a proposal", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSelfRegisteringAgentTestingSuite) {
			proposalAgent := model_builders.NewAgentBuilder().WithoutCollectionTokenId().WithOwner(owner).WithProposalId(9).Build()
			submittedState := model_builders.NewAgentRegistrationStateBuilder().
				WithMintStatus(identity_enums.TransactionStatusSubmitted).
				WithEnsRegistryStatus(identity_enums.TransactionStatusPending).
				Build()

			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(draft(), nil)
			suite.mockAgentRegistrationStateRepository.EXPECT().Get(gomock.Any()).Return(nil, nil)
			suite.mockMintPublisher.EXPECT().SelfMint(gomock.Any(), gomock.Any(), gomock.Any()).Return(
				&identity_interfaces.MintResult{Agent: proposalAgent, State: submittedState, ProposalPending: true}, nil)
		}

		t.Run("Should return submitted and not touch ENS", func(t *testing.T) {
			t.Parallel()

			suite := WhenSelfRegisteringAgentBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, string(identity_enums.TransactionStatusSubmitted), resp.MintStatus)
			assert.NotNil(t, resp.ProposalId)
			assert.Equal(t, uint64(9), *resp.ProposalId)
		})
	})

	t.Run("Given the draft has no owner", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSelfRegisteringAgentTestingSuite) {
			ownerlessDraft := model_builders.NewAgentBuilder().WithoutCollectionTokenId().Build()
			ownerlessDraft.Owner = nil
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(ownerlessDraft, nil)
		}

		t.Run("Should return a Conflict without minting", func(t *testing.T) {
			t.Parallel()

			suite := WhenSelfRegisteringAgentBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background())

			var conflictErr *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictErr)
		})
	})
}
