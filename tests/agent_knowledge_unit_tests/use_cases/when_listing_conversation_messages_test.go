package use_cases_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/helpers"
	identity_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	knowledge_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_knowledge_mocks/repositories_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var (
	listingMessagesCallerId  = strings.ToLower(common_tests.RandomEthAddress().Hex())
	listingMessagesCallerCtx = helpers.ContextWithCallerId(context.Background(), listingMessagesCallerId)
)

type WhenListingConversationMessagesTestingSuite struct {
	sut *use_cases.ListConversationMessages

	mockConversationRepository        *knowledge_mocks.MockConversationRepositoryInterface
	mockConversationMessageRepository *knowledge_mocks.MockConversationMessageRepositoryInterface
	mockAgentRepository               *identity_repositories_mocks.MockAgentRepositoryInterface
}

func WhenListingConversationMessagesBeforeEach(t *testing.T) *WhenListingConversationMessagesTestingSuite {
	mockController := gomock.NewController(t)

	mockConversationRepository := knowledge_mocks.NewMockConversationRepositoryInterface(mockController)
	mockConversationMessageRepository := knowledge_mocks.NewMockConversationMessageRepositoryInterface(mockController)
	mockAgentRepository := identity_repositories_mocks.NewMockAgentRepositoryInterface(mockController)

	return &WhenListingConversationMessagesTestingSuite{
		sut: use_cases.NewListConversationMessages(
			mockConversationRepository,
			mockConversationMessageRepository,
			mockAgentRepository,
		),
		mockConversationRepository:        mockConversationRepository,
		mockConversationMessageRepository: mockConversationMessageRepository,
		mockAgentRepository:               mockAgentRepository,
	}
}

func TestWhenListingConversationMessages(t *testing.T) {
	t.Parallel()

	conversationId := "conv-1"
	request := func() *requests.ListConversationMessagesRequest {
		return &requests.ListConversationMessagesRequest{
			ConversationIdRequest: requests.ConversationIdRequest{ConversationId: conversationId},
		}
	}

	t.Run("Given an invalid request", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationMessagesBeforeEach(t)

			_, err := suite.sut.Execute(listingMessagesCallerCtx, &requests.ListConversationMessagesRequest{
				ConversationIdRequest: requests.ConversationIdRequest{ConversationId: "  "},
			})

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})

	t.Run("Given no caller id", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return Unauthorized without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationMessagesBeforeEach(t)

			_, err := suite.sut.Execute(context.Background(), request())

			var unauthorizedErr *apperrors.UnauthorizedError
			assert.ErrorAs(t, err, &unauthorizedErr)
			assert.Equal(t, "missing_caller_id", unauthorizedErr.Code)
		})
	})

	t.Run("Given the agent lookup fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingConversationMessagesTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(nil, errors.New("db blew"))
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(listingMessagesCallerCtx, request())

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "list_failed", internalErr.Code)
		})
	})

	t.Run("Given the agent identity is not initialized", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingConversationMessagesTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(nil, nil)
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(listingMessagesCallerCtx, request())

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "list_failed", internalErr.Code)
		})
	})

	t.Run("Given the conversation lookup fails", func(t *testing.T) {
		t.Parallel()

		agent := model_builders.NewAgentBuilder().Build()

		initSuite := func(suite *WhenListingConversationMessagesTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockConversationRepository.EXPECT().
				GetByIdAndAgentIdAndCallerId(gomock.Any(), conversationId, agent.Id, lo.ToPtr(listingMessagesCallerId)).
				Return(nil, errors.New("db blew"))
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(listingMessagesCallerCtx, request())

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "list_failed", internalErr.Code)
		})
	})

	t.Run("Given the conversation does not exist", func(t *testing.T) {
		t.Parallel()

		agent := model_builders.NewAgentBuilder().Build()

		initSuite := func(suite *WhenListingConversationMessagesTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockConversationRepository.EXPECT().
				GetByIdAndAgentIdAndCallerId(gomock.Any(), conversationId, agent.Id, lo.ToPtr(listingMessagesCallerId)).
				Return(nil, nil)
		}

		t.Run("Should return NotFoundError", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(listingMessagesCallerCtx, request())

			var notFoundErr *apperrors.NotFoundError
			assert.ErrorAs(t, err, &notFoundErr)
			assert.Equal(t, "conversation_not_found", notFoundErr.Code)
		})
	})

	t.Run("Given listing the messages fails", func(t *testing.T) {
		t.Parallel()

		agent := model_builders.NewAgentBuilder().Build()

		initSuite := func(suite *WhenListingConversationMessagesTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockConversationRepository.EXPECT().
				GetByIdAndAgentIdAndCallerId(gomock.Any(), conversationId, agent.Id, lo.ToPtr(listingMessagesCallerId)).
				Return(model_builders.NewConversationBuilder().WithId(conversationId).WithAgentId(agent.Id).Build(), nil)
			suite.mockConversationMessageRepository.EXPECT().
				ListByConversationId(gomock.Any(), conversationId).
				Return(nil, errors.New("db blew"))
		}

		t.Run("Should return an InternalError carrying the repository message", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(listingMessagesCallerCtx, request())

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.ErrorContains(t, internalErr.Cause, "db blew")
		})
	})

	t.Run("Given the caller does not own the agent", func(t *testing.T) {
		t.Parallel()

		agent := model_builders.NewAgentBuilder().Build()

		initSuite := func(suite *WhenListingConversationMessagesTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockConversationRepository.EXPECT().
				GetByIdAndAgentIdAndCallerId(gomock.Any(), conversationId, agent.Id, lo.ToPtr(listingMessagesCallerId)).
				Return(model_builders.NewConversationBuilder().
					WithId(conversationId).
					WithAgentId(agent.Id).
					WithCallerId(listingMessagesCallerId).
					Build(), nil)
			suite.mockConversationMessageRepository.EXPECT().
				ListByConversationId(gomock.Any(), conversationId).
				Return([]models.ConversationMessage{}, nil)
		}

		t.Run("Should look the conversation up within the caller's scope", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(listingMessagesCallerCtx, request())

			assert.NoError(t, err)
		})
	})

	t.Run("Given the caller owns the agent", func(t *testing.T) {
		t.Parallel()

		agent := model_builders.NewAgentBuilder().WithOwner(strings.ToUpper(listingMessagesCallerId)).Build()

		initSuite := func(suite *WhenListingConversationMessagesTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockConversationRepository.EXPECT().
				GetByIdAndAgentIdAndCallerId(gomock.Any(), conversationId, agent.Id, gomock.Nil()).
				Return(model_builders.NewConversationBuilder().WithId(conversationId).WithAgentId(agent.Id).Build(), nil)
			suite.mockConversationMessageRepository.EXPECT().
				ListByConversationId(gomock.Any(), conversationId).
				Return([]models.ConversationMessage{}, nil)
		}

		t.Run("Should look the conversation up unscoped", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(listingMessagesCallerCtx, request())

			assert.NoError(t, err)
		})
	})

	t.Run("Given the conversation holds messages", func(t *testing.T) {
		t.Parallel()

		agent := model_builders.NewAgentBuilder().Build()
		stored := []models.ConversationMessage{
			model_builders.NewConversationMessageBuilder().
				WithConversationId(conversationId).
				WithTimestamp(time.Now()).
				Build(),
		}

		initSuite := func(suite *WhenListingConversationMessagesTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockConversationRepository.EXPECT().
				GetByIdAndAgentIdAndCallerId(gomock.Any(), conversationId, agent.Id, lo.ToPtr(listingMessagesCallerId)).
				Return(model_builders.NewConversationBuilder().WithId(conversationId).WithAgentId(agent.Id).Build(), nil)
			suite.mockConversationMessageRepository.EXPECT().
				ListByConversationId(gomock.Any(), conversationId).
				Return(stored, nil)
		}

		t.Run("Should return them on the wire", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationMessagesBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(listingMessagesCallerCtx, request())

			assert.NoError(t, err)
			assert.Len(t, resp.Messages, 1)
			assert.Equal(t, conversationId, resp.Messages[0].ConversationId)
		})
	})
}
