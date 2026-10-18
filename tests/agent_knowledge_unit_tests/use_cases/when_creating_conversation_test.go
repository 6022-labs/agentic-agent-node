package use_cases_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/helpers"
	identity_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	knowledge_metrics_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_knowledge_mocks/metrics_mocks/interfaces_mocks"
	knowledge_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_knowledge_mocks/repositories_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var (
	creatingCallerId  = strings.ToLower(common_tests.RandomEthAddress().Hex())
	creatingCallerCtx = helpers.ContextWithCallerId(context.Background(), creatingCallerId)
)

type WhenCreatingConversationTestingSuite struct {
	sut *use_cases.CreateConversation

	mockConversationRepository *knowledge_mocks.MockConversationRepositoryInterface
	mockAgentRepository        *identity_repositories_mocks.MockAgentRepositoryInterface
	mockKnowledgeMetrics       *knowledge_metrics_mocks.MockKnowledgeMetricsInterface
}

func WhenCreatingConversationBeforeEach(t *testing.T) *WhenCreatingConversationTestingSuite {
	mockController := gomock.NewController(t)

	mockConversationRepository := knowledge_mocks.NewMockConversationRepositoryInterface(mockController)
	mockAgentRepository := identity_repositories_mocks.NewMockAgentRepositoryInterface(mockController)
	mockKnowledgeMetrics := knowledge_metrics_mocks.NewMockKnowledgeMetricsInterface(mockController)

	return &WhenCreatingConversationTestingSuite{
		sut: use_cases.NewCreateConversation(
			mockConversationRepository,
			mockAgentRepository,
			mockKnowledgeMetrics,
		),
		mockConversationRepository: mockConversationRepository,
		mockAgentRepository:        mockAgentRepository,
		mockKnowledgeMetrics:       mockKnowledgeMetrics,
	}
}

func TestWhenCreatingConversation(t *testing.T) {
	t.Parallel()

	callerId := creatingCallerId

	t.Run("Given an invalid request", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingConversationBeforeEach(t)

			_, err := suite.sut.Execute(creatingCallerCtx, &requests.CreateConversationRequest{Name: "   "})

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})

	t.Run("Given no caller id", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return Unauthorized without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingConversationBeforeEach(t)

			_, err := suite.sut.Execute(context.Background(), &requests.CreateConversationRequest{Name: "Audit"})

			var unauthorizedErr *apperrors.UnauthorizedError
			assert.ErrorAs(t, err, &unauthorizedErr)
			assert.Equal(t, "missing_caller_id", unauthorizedErr.Code)
		})
	})

	t.Run("Given the agent lookup fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingConversationTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(nil, errors.New("db blew"))
		}

		t.Run("Should return InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingConversationBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(creatingCallerCtx, &requests.CreateConversationRequest{
				Name: "Audit",
			})

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "create_failed", internalErr.Code)
		})
	})

	t.Run("Given the agent identity is not initialized", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingConversationTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(nil, nil)
		}

		t.Run("Should return InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingConversationBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(creatingCallerCtx, &requests.CreateConversationRequest{
				Name: "Audit",
			})

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "create_failed", internalErr.Code)
		})
	})

	t.Run("Given the agent is not minted yet", func(t *testing.T) {
		t.Parallel()

		agent := model_builders.NewAgentBuilder().Build()
		agent.CollectionTokenId = nil

		initSuite := func(suite *WhenCreatingConversationTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
		}

		t.Run("Should return the agent_not_minted conflict", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingConversationBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(creatingCallerCtx, &requests.CreateConversationRequest{
				Name: "whatever",
			})

			var conflictErr *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictErr)
			assert.Equal(t, "agent_not_minted", conflictErr.Code)
		})
	})

	t.Run("Given the name lookup fails", func(t *testing.T) {
		t.Parallel()

		agent := model_builders.NewAgentBuilder().Build()

		initSuite := func(suite *WhenCreatingConversationTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockConversationRepository.EXPECT().
				ExistsByAgentIdAndCallerIdAndName(gomock.Any(), agent.Id, &callerId, "Audit", nil).
				Return(false, errors.New("db blew"))
		}

		t.Run("Should return InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingConversationBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(creatingCallerCtx, &requests.CreateConversationRequest{
				Name: "Audit",
			})

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "create_failed", internalErr.Code)
		})
	})

	t.Run("Given the name is already taken", func(t *testing.T) {
		t.Parallel()

		agent := model_builders.NewAgentBuilder().Build()

		initSuite := func(suite *WhenCreatingConversationTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockConversationRepository.EXPECT().
				ExistsByAgentIdAndCallerIdAndName(gomock.Any(), agent.Id, &callerId, "taken", nil).
				Return(true, nil)
		}

		t.Run("Should return a ConflictError", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingConversationBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(creatingCallerCtx, &requests.CreateConversationRequest{
				Name: "taken",
			})

			var conflictErr *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictErr)
			assert.Equal(t, "conversation_name_taken", conflictErr.Code)
		})
	})

	t.Run("Given the store fails", func(t *testing.T) {
		t.Parallel()

		agent := model_builders.NewAgentBuilder().Build()

		initSuite := func(suite *WhenCreatingConversationTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockConversationRepository.EXPECT().
				ExistsByAgentIdAndCallerIdAndName(gomock.Any(), agent.Id, &callerId, "Audit", nil).
				Return(false, nil)
			suite.mockConversationRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(nil, errors.New("db blew"))
		}

		t.Run("Should return InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingConversationBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(creatingCallerCtx, &requests.CreateConversationRequest{
				Name: "Audit",
			})

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "create_failed", internalErr.Code)
		})
	})

	t.Run("Given a free name", func(t *testing.T) {
		t.Parallel()

		agent := model_builders.NewAgentBuilder().Build()

		initSuite := func(suite *WhenCreatingConversationTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockConversationRepository.EXPECT().
				ExistsByAgentIdAndCallerIdAndName(gomock.Any(), agent.Id, &callerId, "ERC20 mint audit", nil).
				Return(false, nil)
			suite.mockConversationRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, conversation *models.Conversation) (*models.Conversation, error) {
					_, err := uuid.Parse(conversation.Id)
					assert.NoError(t, err, "the id must be server-generated")
					assert.Equal(t, agent.Id, conversation.AgentId)
					assert.Equal(t, "ERC20 mint audit", *conversation.Name)
					assert.Equal(t, callerId, *conversation.CallerId)
					return conversation, nil
				})
			suite.mockKnowledgeMetrics.EXPECT().RecordConversationStarted(gomock.Any())
			suite.mockKnowledgeMetrics.EXPECT().RecordConversationNamed(gomock.Any())
		}

		t.Run("Should create a caller-owned session with a generated id", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingConversationBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(creatingCallerCtx, &requests.CreateConversationRequest{
				Name: "  ERC20   mint audit ",
			})

			assert.NoError(t, err)
			assert.Equal(t, "ERC20 mint audit", *resp.Name)
			assert.NotEmpty(t, resp.ConversationId)
		})
	})

	t.Run("Given the caller owns the agent", func(t *testing.T) {
		t.Parallel()

		agent := model_builders.NewAgentBuilder().WithOwner(strings.ToUpper(callerId)).Build()

		initSuite := func(suite *WhenCreatingConversationTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockConversationRepository.EXPECT().
				ExistsByAgentIdAndCallerIdAndName(gomock.Any(), agent.Id, &callerId, "Audit", nil).
				Return(false, nil)
			suite.mockConversationRepository.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, conversation *models.Conversation) (*models.Conversation, error) {
					assert.Equal(t, callerId, *conversation.CallerId)
					return conversation, nil
				})
			suite.mockKnowledgeMetrics.EXPECT().RecordConversationStarted(gomock.Any())
			suite.mockKnowledgeMetrics.EXPECT().RecordConversationNamed(gomock.Any())
		}

		t.Run("Should still own the session as its caller", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingConversationBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(creatingCallerCtx, &requests.CreateConversationRequest{Name: "Audit"})

			assert.NoError(t, err)
		})
	})
}
