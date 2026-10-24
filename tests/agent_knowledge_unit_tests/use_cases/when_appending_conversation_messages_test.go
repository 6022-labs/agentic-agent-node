package use_cases_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	identity_models "github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/helpers"
	identity_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	knowledge_metrics_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_knowledge_mocks/metrics_mocks/interfaces_mocks"
	knowledge_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_knowledge_mocks/repositories_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/agent_knowledge_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var (
	appendingCallerId  = strings.ToLower(common_tests.RandomEthAddress().Hex())
	appendingCallerCtx = helpers.ContextWithCallerId(context.Background(), appendingCallerId)
)

type WhenAppendingConversationMessagesTestingSuite struct {
	sut *use_cases.AppendConversationMessages

	mockMessageAppender        *interfaces_mocks.MockMessageAppenderInterface
	mockConversationRepository *knowledge_repositories_mocks.MockConversationRepositoryInterface
	mockAgentRepository        *identity_repositories_mocks.MockAgentRepositoryInterface
	mockKnowledgeMetrics       *knowledge_metrics_mocks.MockKnowledgeMetricsInterface
}

func WhenAppendingConversationMessagesBeforeEach(t *testing.T) *WhenAppendingConversationMessagesTestingSuite {
	mockController := gomock.NewController(t)

	mockMessageAppender := interfaces_mocks.NewMockMessageAppenderInterface(mockController)
	mockConversationRepository := knowledge_repositories_mocks.NewMockConversationRepositoryInterface(mockController)
	mockAgentRepository := identity_repositories_mocks.NewMockAgentRepositoryInterface(mockController)
	mockKnowledgeMetrics := knowledge_metrics_mocks.NewMockKnowledgeMetricsInterface(mockController)

	sut := use_cases.NewAppendConversationMessages(
		mockMessageAppender,
		mockConversationRepository,
		mockAgentRepository,
		mockKnowledgeMetrics,
	)

	return &WhenAppendingConversationMessagesTestingSuite{
		sut:                        sut,
		mockMessageAppender:        mockMessageAppender,
		mockConversationRepository: mockConversationRepository,
		mockAgentRepository:        mockAgentRepository,
		mockKnowledgeMetrics:       mockKnowledgeMetrics,
	}
}

func (s *WhenAppendingConversationMessagesTestingSuite) expectMintedAgent() *identity_models.Agent {
	agent := model_builders.NewAgentBuilder().Build()
	s.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil).Times(1)
	return agent
}

func (s *WhenAppendingConversationMessagesTestingSuite) expectReachableConversation(
	agent *identity_models.Agent,
	conversationId string,
) {
	s.mockConversationRepository.EXPECT().Ensure(gomock.Any(), gomock.Any()).Return(false, nil)
	s.mockConversationRepository.EXPECT().
		GetByIdAndAgentIdAndCallerId(gomock.Any(), conversationId, agent.Id, lo.ToPtr(appendingCallerId)).
		Return(model_builders.NewConversationBuilder().
			WithId(conversationId).
			WithAgentId(agent.Id).
			WithCallerId(appendingCallerId).
			Build(), nil)
}

func TestWhenAppendingConversationMessages(t *testing.T) {
	t.Parallel()

	now := time.Now()
	conversationId := "conv-1"

	t.Run("Given an invalid request", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)

			_, err := suite.sut.Execute(
				appendingCallerCtx,
				&requests.AppendConversationMessagesRequest{
					ConversationIdRequest: requests.ConversationIdRequest{ConversationId: conversationId},
				},
			)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})

	t.Run("Given no caller id", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return Unauthorized without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)

			_, err := suite.sut.Execute(
				context.Background(),
				&requests.AppendConversationMessagesRequest{
					ConversationIdRequest: requests.ConversationIdRequest{ConversationId: conversationId},
					Messages: []requests.ConversationMessageInput{
						{Role: "user", Content: "hi", Timestamp: now},
					},
				},
			)

			var unauthorizedErr *apperrors.UnauthorizedError
			assert.ErrorAs(t, err, &unauthorizedErr)
			assert.Equal(t, "missing_caller_id", unauthorizedErr.Code)
		})
	})

	t.Run("Given the agent lookup fails", func(t *testing.T) {
		t.Parallel()

		req := &requests.AppendConversationMessagesRequest{
			ConversationIdRequest: requests.ConversationIdRequest{ConversationId: conversationId},
			Messages: []requests.ConversationMessageInput{
				{Role: "user", Content: "hi", Timestamp: now},
			},
		}

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(nil, errors.New("db down")).Times(1)
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(appendingCallerCtx, req)

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "append_failed", internalErr.Code)
		})
	})

	t.Run("Given the agent identity is not initialized", func(t *testing.T) {
		t.Parallel()

		req := &requests.AppendConversationMessagesRequest{
			ConversationIdRequest: requests.ConversationIdRequest{ConversationId: conversationId},
			Messages: []requests.ConversationMessageInput{
				{Role: "user", Content: "hi", Timestamp: now},
			},
		}

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(nil, nil).Times(1)
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(appendingCallerCtx, req)

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "append_failed", internalErr.Code)
		})
	})

	t.Run("Given the agent is not minted yet", func(t *testing.T) {
		t.Parallel()

		req := &requests.AppendConversationMessagesRequest{
			ConversationIdRequest: requests.ConversationIdRequest{ConversationId: conversationId},
			Messages: []requests.ConversationMessageInput{
				{Role: "user", Content: "hi", Timestamp: now},
			},
		}

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			agent := model_builders.NewAgentBuilder().WithoutCollectionTokenId().Build()
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil).Times(1)
		}

		t.Run("Should return the agent_not_minted conflict", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(appendingCallerCtx, req)

			var conflictErr *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictErr)
			assert.Equal(t, "agent_not_minted", conflictErr.Code)
		})
	})

	t.Run("Given creating the conversation fails", func(t *testing.T) {
		t.Parallel()

		req := &requests.AppendConversationMessagesRequest{
			ConversationIdRequest: requests.ConversationIdRequest{ConversationId: conversationId},
			Messages: []requests.ConversationMessageInput{
				{Role: "user", Content: "hi", Timestamp: now},
			},
		}

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			suite.expectMintedAgent()
			suite.mockConversationRepository.EXPECT().
				Ensure(gomock.Any(), gomock.Any()).
				Return(false, errors.New("db down"))
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(appendingCallerCtx, req)

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "append_failed", internalErr.Code)
		})
	})

	t.Run("Given the conversation lookup fails", func(t *testing.T) {
		t.Parallel()

		req := &requests.AppendConversationMessagesRequest{
			ConversationIdRequest: requests.ConversationIdRequest{ConversationId: conversationId},
			Messages: []requests.ConversationMessageInput{
				{Role: "user", Content: "hi", Timestamp: now},
			},
		}

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			agent := suite.expectMintedAgent()
			suite.mockConversationRepository.EXPECT().Ensure(gomock.Any(), gomock.Any()).Return(false, nil)
			suite.mockConversationRepository.EXPECT().
				GetByIdAndAgentIdAndCallerId(gomock.Any(), conversationId, agent.Id, lo.ToPtr(appendingCallerId)).
				Return(nil, errors.New("db down"))
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(appendingCallerCtx, req)

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "append_failed", internalErr.Code)
		})
	})

	t.Run("Given another caller owns the conversation", func(t *testing.T) {
		t.Parallel()

		req := &requests.AppendConversationMessagesRequest{
			ConversationIdRequest: requests.ConversationIdRequest{ConversationId: conversationId},
			Messages: []requests.ConversationMessageInput{
				{Role: "user", Content: "hi", Timestamp: now},
			},
		}

		t.Run("Given the caller does not own the agent", func(t *testing.T) {
			t.Parallel()

			initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
				agent := suite.expectMintedAgent()
				suite.mockConversationRepository.EXPECT().Ensure(gomock.Any(), gomock.Any()).Return(false, nil)
				suite.mockConversationRepository.EXPECT().
					GetByIdAndAgentIdAndCallerId(gomock.Any(), conversationId, agent.Id, lo.ToPtr(appendingCallerId)).
					Return(nil, nil)
				suite.mockMessageAppender.EXPECT().Append(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			}

			t.Run("Should return NotFoundError without appending", func(t *testing.T) {
				t.Parallel()

				suite := WhenAppendingConversationMessagesBeforeEach(t)
				initSuite(suite)

				_, err := suite.sut.Execute(appendingCallerCtx, req)

				var notFoundErr *apperrors.NotFoundError
				assert.ErrorAs(t, err, &notFoundErr)
				assert.Equal(t, "conversation_not_found", notFoundErr.Code)
			})
		})

		t.Run("Given the caller owns the agent", func(t *testing.T) {
			t.Parallel()

			initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
				agent := model_builders.NewAgentBuilder().WithOwner(strings.ToUpper(appendingCallerId)).Build()
				suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
				suite.mockConversationRepository.EXPECT().Ensure(gomock.Any(), gomock.Any()).Return(false, nil)
				suite.mockConversationRepository.EXPECT().
					GetByIdAndAgentIdAndCallerId(gomock.Any(), conversationId, agent.Id, gomock.Nil()).
					Return(model_builders.NewConversationBuilder().
						WithId(conversationId).
						WithAgentId(agent.Id).
						WithCallerId(strings.ToLower(common_tests.RandomEthAddress().Hex())).
						Build(), nil)
				suite.mockMessageAppender.EXPECT().Append(gomock.Any(), conversationId, gomock.Any()).Return(nil, nil)
			}

			t.Run("Should seed it", func(t *testing.T) {
				t.Parallel()

				suite := WhenAppendingConversationMessagesBeforeEach(t)
				initSuite(suite)

				_, err := suite.sut.Execute(appendingCallerCtx, req)

				assert.NoError(t, err)
			})
		})
	})

	t.Run("Given the message appender fails", func(t *testing.T) {
		t.Parallel()

		req := &requests.AppendConversationMessagesRequest{
			ConversationIdRequest: requests.ConversationIdRequest{ConversationId: conversationId},
			Messages: []requests.ConversationMessageInput{
				{Role: "user", Timestamp: now},
			},
		}
		mapped := []models.ConversationMessage{
			{Role: "user", Timestamp: now},
		}

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			agent := suite.expectMintedAgent()
			suite.expectReachableConversation(agent, conversationId)
			suite.mockMessageAppender.EXPECT().
				Append(gomock.Any(), conversationId, mapped).
				Return(nil, errors.New("db down"))
		}

		t.Run("Should return an InternalError carrying the appender message", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(appendingCallerCtx, req)

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Contains(t, internalErr.Message, "db down")
		})
	})

	t.Run("Given the conversation is new", func(t *testing.T) {
		t.Parallel()

		req := &requests.AppendConversationMessagesRequest{
			ConversationIdRequest: requests.ConversationIdRequest{ConversationId: conversationId},
			Messages: []requests.ConversationMessageInput{
				{Role: "user", Content: "hi", Timestamp: now},
			},
		}

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			agent := suite.expectMintedAgent()
			gomock.InOrder(
				suite.mockConversationRepository.EXPECT().
					Ensure(gomock.Any(), &models.Conversation{
						Id:       conversationId,
						AgentId:  agent.Id,
						CallerId: lo.ToPtr(appendingCallerId),
					}).
					Return(true, nil),
				suite.mockConversationRepository.EXPECT().
					GetByIdAndAgentIdAndCallerId(gomock.Any(), conversationId, agent.Id, lo.ToPtr(appendingCallerId)).
					Return(model_builders.NewConversationBuilder().
						WithId(conversationId).
						WithAgentId(agent.Id).
						WithCallerId(appendingCallerId).
						Build(), nil),
				suite.mockMessageAppender.EXPECT().
					Append(gomock.Any(), conversationId, gomock.Any()).
					Return(nil, nil),
			)
			suite.mockKnowledgeMetrics.EXPECT().RecordConversationStarted(gomock.Any())
		}

		t.Run("Should create it for the caller", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(appendingCallerCtx, req)

			assert.NoError(t, err)
		})
	})

	t.Run("Given the conversation already exists", func(t *testing.T) {
		t.Parallel()

		req := &requests.AppendConversationMessagesRequest{
			ConversationIdRequest: requests.ConversationIdRequest{ConversationId: conversationId},
			Messages: []requests.ConversationMessageInput{
				{Role: "user", Content: "hi", Timestamp: now},
			},
		}

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			agent := suite.expectMintedAgent()
			suite.expectReachableConversation(agent, conversationId)
			suite.mockMessageAppender.EXPECT().Append(gomock.Any(), conversationId, gomock.Any()).Return(nil, nil)
		}

		t.Run("Should not count it as started", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(appendingCallerCtx, req)

			assert.NoError(t, err)
		})
	})

	t.Run("Given the message appender persists the messages", func(t *testing.T) {
		t.Parallel()

		req := &requests.AppendConversationMessagesRequest{
			ConversationIdRequest: requests.ConversationIdRequest{ConversationId: conversationId},
			Messages: []requests.ConversationMessageInput{
				{Role: "user", Content: "hi", Timestamp: now},
			},
		}
		mapped := []models.ConversationMessage{
			{Role: "user", Content: lo.EmptyableToPtr("hi"), Timestamp: now},
		}

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			agent := suite.expectMintedAgent()
			suite.expectReachableConversation(agent, conversationId)
			suite.mockMessageAppender.EXPECT().Append(gomock.Any(), conversationId, mapped).Return(mapped, nil)
		}

		t.Run("Should return the persisted rows", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(appendingCallerCtx, req)

			assert.NoError(t, err)
			assert.Len(t, resp.Inserted, 1)
			assert.Equal(t, "user", resp.Inserted[0].Role)
		})
	})
}
