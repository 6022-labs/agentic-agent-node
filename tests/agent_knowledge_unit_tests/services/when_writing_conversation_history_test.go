package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/services"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/services/interfaces"
	knowledge_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_knowledge_mocks/repositories_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/agent_knowledge_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenWritingConversationHistoryTestingSuite struct {
	sut *services.ConversationHistory

	mockRepository      *knowledge_mocks.MockConversationMessageRepositoryInterface
	mockMessageAppender *interfaces_mocks.MockMessageAppenderInterface
}

func WhenWritingConversationHistoryBeforeEach(t *testing.T) *WhenWritingConversationHistoryTestingSuite {
	mockController := gomock.NewController(t)

	mockRepository := knowledge_mocks.NewMockConversationMessageRepositoryInterface(mockController)
	mockMessageAppender := interfaces_mocks.NewMockMessageAppenderInterface(mockController)

	sut := services.NewConversationHistory(mockRepository, mockMessageAppender)

	return &WhenWritingConversationHistoryTestingSuite{
		sut:                 sut,
		mockRepository:      mockRepository,
		mockMessageAppender: mockMessageAppender,
	}
}

func TestWhenWritingConversationHistory(t *testing.T) {
	t.Parallel()

	conversationId := "conv-1"
	now := time.Now().UTC().Truncate(time.Microsecond)

	t.Run("Given DTO inputs", func(t *testing.T) {
		t.Parallel()

		inputs := []interfaces.ConversationMessage{
			{Role: "user", Content: "hi", Timestamp: now},
			{Role: "assistant", Content: "hello", Timestamp: now},
		}

		initSuite := func(suite *WhenWritingConversationHistoryTestingSuite) {
			suite.mockMessageAppender.EXPECT().Append(gomock.Any(),
				conversationId, gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, msgs []models.ConversationMessage) ([]models.ConversationMessage, error) {
					if assert.Len(t, msgs, 2) {
						assert.Equal(t, "user", msgs[0].Role.String())
						assert.Equal(t, "hi", lo.FromPtr(msgs[0].Content))
						assert.Equal(t, now, msgs[0].Timestamp)
						assert.Equal(t, "assistant", msgs[1].Role.String())
					}
					return msgs, nil
				})
		}

		t.Run("Should map DTOs to persistence models and delegate to the message appender", func(t *testing.T) {
			t.Parallel()

			suite := WhenWritingConversationHistoryBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Write(context.Background(), conversationId, inputs)
			assert.NoError(t, err)
			assert.Len(t, result, 2)
		})
	})

	t.Run("Given the message appender returns persistence rows", func(t *testing.T) {
		t.Parallel()

		persisted := []models.ConversationMessage{
			model_builders.NewConversationMessageBuilder().
				WithRole("user").
				WithContent("hi").
				WithConversationId(conversationId).
				WithTimestamp(now).
				Build(),
		}

		initSuite := func(suite *WhenWritingConversationHistoryTestingSuite) {
			suite.mockMessageAppender.EXPECT().Append(gomock.Any(),
				conversationId, gomock.Any()).
				Return(persisted, nil)
		}

		t.Run("Should map them back to DTOs (no Hash, no ConversationId leak)", func(t *testing.T) {
			t.Parallel()

			suite := WhenWritingConversationHistoryBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Write(context.Background(), conversationId, []interfaces.ConversationMessage{
				{Role: "user", Content: "hi", Timestamp: now},
			})

			assert.NoError(t, err)
			if assert.Len(t, result, 1) {
				assert.Equal(t, "hi", result[0].Content)
				assert.Equal(t, now, result[0].Timestamp)
			}
		})
	})

	t.Run("Given the message appender fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenWritingConversationHistoryTestingSuite) {
			suite.mockMessageAppender.EXPECT().Append(gomock.Any(),
				conversationId, gomock.Any()).
				Return(nil, errors.New("db down"))
		}

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenWritingConversationHistoryBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Write(context.Background(), conversationId, []interfaces.ConversationMessage{
				{Role: "user", Content: "hi"},
			})

			assert.ErrorContains(t, err, "db down")
		})
	})
}
