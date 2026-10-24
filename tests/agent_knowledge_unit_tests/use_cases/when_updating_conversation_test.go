package use_cases_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/helpers"
	identity_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	knowledge_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_knowledge_mocks/repositories_mocks"
	knowledge_services_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_knowledge_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var (
	updatingCallerId  = strings.ToLower(common_tests.RandomEthAddress().Hex())
	updatingCallerCtx = helpers.ContextWithCallerId(context.Background(), updatingCallerId)
)

type WhenUpdatingConversationTestingSuite struct {
	sut *use_cases.UpdateConversation

	mockConversationRepository *knowledge_mocks.MockConversationRepositoryInterface
	mockConversationLabeler    *knowledge_services_mocks.MockConversationLabelerInterface
	mockAgentRepository        *identity_repositories_mocks.MockAgentRepositoryInterface
}

func WhenUpdatingConversationBeforeEach(t *testing.T) *WhenUpdatingConversationTestingSuite {
	mockController := gomock.NewController(t)

	mockConversationRepository := knowledge_mocks.NewMockConversationRepositoryInterface(mockController)
	mockConversationLabeler := knowledge_services_mocks.NewMockConversationLabelerInterface(mockController)
	mockAgentRepository := identity_repositories_mocks.NewMockAgentRepositoryInterface(mockController)

	return &WhenUpdatingConversationTestingSuite{
		sut: use_cases.NewUpdateConversation(
			mockConversationRepository,
			mockConversationLabeler,
			mockAgentRepository,
		),
		mockConversationRepository: mockConversationRepository,
		mockConversationLabeler:    mockConversationLabeler,
		mockAgentRepository:        mockAgentRepository,
	}
}

func TestWhenUpdatingConversation(t *testing.T) {
	t.Parallel()

	t.Run("Given an invalid request", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingConversationBeforeEach(t)

			_, err := suite.sut.Execute(updatingCallerCtx, &requests.UpdateConversationRequest{
				ConversationIdRequest:   requests.ConversationIdRequest{ConversationId: "c1"},
				ConversationNameRequest: requests.ConversationNameRequest{Name: "   "},
			})

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})

	t.Run("Given no caller id", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return Unauthorized without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingConversationBeforeEach(t)

			_, err := suite.sut.Execute(context.Background(), &requests.UpdateConversationRequest{
				ConversationIdRequest:   requests.ConversationIdRequest{ConversationId: "c1"},
				ConversationNameRequest: requests.ConversationNameRequest{Name: "Audit"},
			})

			var unauthorizedErr *apperrors.UnauthorizedError
			assert.ErrorAs(t, err, &unauthorizedErr)
			assert.Equal(t, "missing_caller_id", unauthorizedErr.Code)
		})
	})

	t.Run("Given the agent lookup fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingConversationTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(nil, errors.New("db blew"))
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingConversationBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(updatingCallerCtx, &requests.UpdateConversationRequest{
				ConversationIdRequest:   requests.ConversationIdRequest{ConversationId: "c1"},
				ConversationNameRequest: requests.ConversationNameRequest{Name: "Audit"},
			})

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "update_failed", internalErr.Code)
		})
	})

	t.Run("Given the agent identity is not initialized", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingConversationTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(nil, nil)
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingConversationBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(updatingCallerCtx, &requests.UpdateConversationRequest{
				ConversationIdRequest:   requests.ConversationIdRequest{ConversationId: "c1"},
				ConversationNameRequest: requests.ConversationNameRequest{Name: "Audit"},
			})

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "update_failed", internalErr.Code)
		})
	})

	t.Run("Given the conversation lookup fails", func(t *testing.T) {
		t.Parallel()

		agent := model_builders.NewAgentBuilder().Build()

		initSuite := func(suite *WhenUpdatingConversationTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockConversationRepository.EXPECT().
				GetByIdAndAgentIdAndCallerId(gomock.Any(), "c1", agent.Id, lo.ToPtr(updatingCallerId)).
				Return(nil, errors.New("db blew"))
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingConversationBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(updatingCallerCtx, &requests.UpdateConversationRequest{
				ConversationIdRequest:   requests.ConversationIdRequest{ConversationId: "c1"},
				ConversationNameRequest: requests.ConversationNameRequest{Name: "mine"},
			})

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "update_failed", internalErr.Code)
		})
	})

	t.Run("Given the conversation does not exist", func(t *testing.T) {
		t.Parallel()

		agent := model_builders.NewAgentBuilder().Build()

		initSuite := func(suite *WhenUpdatingConversationTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockConversationRepository.EXPECT().
				GetByIdAndAgentIdAndCallerId(gomock.Any(), "c1", agent.Id, lo.ToPtr(updatingCallerId)).
				Return(nil, nil)
		}

		t.Run("Should return NotFoundError", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingConversationBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(updatingCallerCtx, &requests.UpdateConversationRequest{
				ConversationIdRequest:   requests.ConversationIdRequest{ConversationId: "c1"},
				ConversationNameRequest: requests.ConversationNameRequest{Name: "mine"},
			})

			var notFoundErr *apperrors.NotFoundError
			assert.ErrorAs(t, err, &notFoundErr)
			assert.Equal(t, "conversation_not_found", notFoundErr.Code)
		})
	})

	t.Run("Given another conversation holds that name", func(t *testing.T) {
		t.Parallel()

		agent := model_builders.NewAgentBuilder().Build()
		conversation := model_builders.NewConversationBuilder().WithId("c1").WithAgentId(agent.Id).Build()

		initSuite := func(suite *WhenUpdatingConversationTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockConversationRepository.EXPECT().
				GetByIdAndAgentIdAndCallerId(gomock.Any(), "c1", agent.Id, lo.ToPtr(updatingCallerId)).
				Return(conversation, nil)
			suite.mockConversationLabeler.EXPECT().
				SetName(gomock.Any(), "c1", "taken").
				Return(nil, interfaces.ErrConversationNameTaken)
		}

		t.Run("Should return a ConflictError", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingConversationBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(updatingCallerCtx, &requests.UpdateConversationRequest{
				ConversationIdRequest:   requests.ConversationIdRequest{ConversationId: "c1"},
				ConversationNameRequest: requests.ConversationNameRequest{Name: "taken"},
			})

			var conflictErr *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictErr)
			assert.Equal(t, "conversation_name_taken", conflictErr.Code)
		})
	})

	t.Run("Given saving the name fails", func(t *testing.T) {
		t.Parallel()

		agent := model_builders.NewAgentBuilder().Build()
		conversation := model_builders.NewConversationBuilder().WithId("c1").WithAgentId(agent.Id).Build()

		initSuite := func(suite *WhenUpdatingConversationTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockConversationRepository.EXPECT().
				GetByIdAndAgentIdAndCallerId(gomock.Any(), "c1", agent.Id, lo.ToPtr(updatingCallerId)).
				Return(conversation, nil)
			suite.mockConversationLabeler.EXPECT().
				SetName(gomock.Any(), "c1", "Audit").
				Return(nil, errors.New("db blew"))
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingConversationBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(updatingCallerCtx, &requests.UpdateConversationRequest{
				ConversationIdRequest:   requests.ConversationIdRequest{ConversationId: "c1"},
				ConversationNameRequest: requests.ConversationNameRequest{Name: "Audit"},
			})

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "update_failed", internalErr.Code)
		})
	})

	t.Run("Given the caller does not own the agent", func(t *testing.T) {
		t.Parallel()

		agent := model_builders.NewAgentBuilder().Build()
		conversation := model_builders.NewConversationBuilder().
			WithId("c1").
			WithAgentId(agent.Id).
			WithCallerId(updatingCallerId).
			Build()

		initSuite := func(suite *WhenUpdatingConversationTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockConversationRepository.EXPECT().
				GetByIdAndAgentIdAndCallerId(gomock.Any(), "c1", agent.Id, lo.ToPtr(updatingCallerId)).
				Return(conversation, nil)
			suite.mockConversationLabeler.EXPECT().
				SetName(gomock.Any(), "c1", "Audit").
				Return(model_builders.NewConversationBuilder().WithId("c1").WithName("Audit").Build(), nil)
		}

		t.Run("Should look the conversation up within the caller's scope", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingConversationBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(updatingCallerCtx, &requests.UpdateConversationRequest{
				ConversationIdRequest:   requests.ConversationIdRequest{ConversationId: "c1"},
				ConversationNameRequest: requests.ConversationNameRequest{Name: "Audit"},
			})

			assert.NoError(t, err)
			assert.Equal(t, "Audit", *resp.Name)
		})
	})

	t.Run("Given the caller owns the agent", func(t *testing.T) {
		t.Parallel()

		agent := model_builders.NewAgentBuilder().WithOwner(strings.ToUpper(updatingCallerId)).Build()
		conversation := model_builders.NewConversationBuilder().WithId("c1").WithAgentId(agent.Id).Build()

		initSuite := func(suite *WhenUpdatingConversationTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockConversationRepository.EXPECT().
				GetByIdAndAgentIdAndCallerId(gomock.Any(), "c1", agent.Id, gomock.Nil()).
				Return(conversation, nil)
			suite.mockConversationLabeler.EXPECT().
				SetName(gomock.Any(), "c1", "Solidity audits").
				Return(model_builders.NewConversationBuilder().
					WithId("c1").
					WithAgentId(agent.Id).
					WithName("Solidity audits").
					Build(), nil)
		}

		t.Run("Should look the conversation up unscoped", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingConversationBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(updatingCallerCtx, &requests.UpdateConversationRequest{
				ConversationIdRequest:   requests.ConversationIdRequest{ConversationId: "c1"},
				ConversationNameRequest: requests.ConversationNameRequest{Name: "Solidity audits"},
			})

			assert.NoError(t, err)
			assert.Equal(t, "Solidity audits", *resp.Name)
		})
	})

	t.Run("Given the new name is the current one", func(t *testing.T) {
		t.Parallel()

		agent := model_builders.NewAgentBuilder().Build()
		conversation := model_builders.NewConversationBuilder().
			WithId("c1").
			WithAgentId(agent.Id).
			WithName("same").
			Build()

		initSuite := func(suite *WhenUpdatingConversationTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockConversationRepository.EXPECT().
				GetByIdAndAgentIdAndCallerId(gomock.Any(), "c1", agent.Id, lo.ToPtr(updatingCallerId)).
				Return(conversation, nil)
			suite.mockConversationLabeler.EXPECT().
				SetName(gomock.Any(), "c1", "same").
				Return(conversation, nil)
		}

		t.Run("Should return the unchanged name", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingConversationBeforeEach(t)
			initSuite(suite)

			updated, err := suite.sut.Execute(updatingCallerCtx, &requests.UpdateConversationRequest{
				ConversationIdRequest:   requests.ConversationIdRequest{ConversationId: "c1"},
				ConversationNameRequest: requests.ConversationNameRequest{Name: "same"},
			})

			assert.NoError(t, err)
			assert.Equal(t, "same", lo.FromPtr(updated.Name))
		})
	})
}
