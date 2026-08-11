package use_cases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	identity_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/agent_knowledge_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenAppendingConversationMessagesTestingSuite struct {
	sut *use_cases.AppendConversationMessages

	mockMessageAppender *interfaces_mocks.MockMessageAppenderInterface
	mockAgentRepository *identity_repositories_mocks.MockAgentRepositoryInterface
}

func WhenAppendingConversationMessagesBeforeEach(t *testing.T) *WhenAppendingConversationMessagesTestingSuite {
	mockController := gomock.NewController(t)

	mockMessageAppender := interfaces_mocks.NewMockMessageAppenderInterface(mockController)
	mockAgentRepository := identity_repositories_mocks.NewMockAgentRepositoryInterface(mockController)

	sut := use_cases.NewAppendConversationMessages(mockMessageAppender, mockAgentRepository)

	return &WhenAppendingConversationMessagesTestingSuite{
		sut:                 sut,
		mockMessageAppender: mockMessageAppender,
		mockAgentRepository: mockAgentRepository,
	}
}

func expectMintedAgent(suite *WhenAppendingConversationMessagesTestingSuite) {
	agent := model_builders.NewAgentBuilder().Build()
	suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil).Times(1)
}

func TestWhenAppendingConversationMessages(t *testing.T) {
	t.Parallel()

	now := time.Now()
	conversationId := "conv-1"

	t.Run("Given the request is invalid", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)

			_, err := suite.sut.Execute(
				context.Background(),
				conversationId,
				&requests.AppendConversationMessagesRequest{},
			)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})

	t.Run("Given the agent identity is not initialized", func(t *testing.T) {
		t.Parallel()

		req := &requests.AppendConversationMessagesRequest{
			Messages: []requests.ConversationMessageInput{
				{Role: "user", Content: "hi", Timestamp: now},
			},
		}

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(nil, nil).Times(1)
		}

		t.Run("Should return an internal error without appending", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), conversationId, req)

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
		})
	})

	t.Run("Given the agent is not minted yet", func(t *testing.T) {
		t.Parallel()

		req := &requests.AppendConversationMessagesRequest{
			Messages: []requests.ConversationMessageInput{
				{Role: "user", Content: "hi", Timestamp: now},
			},
		}

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			agent := model_builders.NewAgentBuilder().WithoutCollectionTokenId().Build()
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil).Times(1)
		}

		t.Run("Should return the agent_not_minted conflict without appending", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), conversationId, req)

			var conflictErr *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictErr)
			assert.Equal(t, "agent_not_minted", conflictErr.Code)
		})
	})

	t.Run("Given the message appender persists the messages", func(t *testing.T) {
		t.Parallel()

		req := &requests.AppendConversationMessagesRequest{
			Messages: []requests.ConversationMessageInput{
				{Role: "user", Content: "hi", Timestamp: now},
			},
		}
		mapped := []models.ConversationMessage{
			{Role: "user", Content: lo.EmptyableToPtr("hi"), Timestamp: now},
		}

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			expectMintedAgent(suite)
			suite.mockMessageAppender.EXPECT().Append(gomock.Any(), conversationId, mapped).Return(mapped, nil)
		}

		t.Run("Should return Ok with the persisted rows on the wire", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background(), conversationId, req)

			assert.NoError(t, err)
			assert.Len(t, resp.Inserted, 1)
			assert.Equal(t, "user", resp.Inserted[0].Role)
		})
	})

	t.Run("Given the message appender fails", func(t *testing.T) {
		t.Parallel()

		req := &requests.AppendConversationMessagesRequest{
			Messages: []requests.ConversationMessageInput{
				{Role: "user", Timestamp: now},
			},
		}
		mapped := []models.ConversationMessage{
			{Role: "user", Timestamp: now},
		}

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			expectMintedAgent(suite)
			suite.mockMessageAppender.EXPECT().
				Append(gomock.Any(), conversationId, mapped).
				Return(nil, errors.New("db down"))
		}

		t.Run("Should return InternalError carrying the appender message", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), conversationId, req)

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Contains(t, internalErr.Message, "db down")
		})
	})
}
