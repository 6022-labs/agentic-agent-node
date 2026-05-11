package services_test

import (
	"errors"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/services"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/tests/agent_knowledge_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenWritingConversationHistoryTestingSuite struct {
	sut *services.ConversationHistory

	mockListService   *interfaces_mocks.MockListConversationMessagesInterface
	mockAppendService *interfaces_mocks.MockAppendConversationMessagesInterface
}

func WhenWritingConversationHistoryBeforeEach(t *testing.T) *WhenWritingConversationHistoryTestingSuite {
	mockController := gomock.NewController(t)

	mockListService := interfaces_mocks.NewMockListConversationMessagesInterface(mockController)
	mockAppendService := interfaces_mocks.NewMockAppendConversationMessagesInterface(mockController)

	sut := services.NewConversationHistory(mockListService, mockAppendService)

	return &WhenWritingConversationHistoryTestingSuite{
		sut:               sut,
		mockListService:   mockListService,
		mockAppendService: mockAppendService,
	}
}

func TestWhenWritingConversationHistory(t *testing.T) {
	t.Parallel()

	conversationId := "conv-1"
	now := time.Now().UTC().Truncate(time.Microsecond)

	t.Run("Given DTO inputs", func(t *testing.T) {
		t.Parallel()

		inputs := []interfaces.ConversationMessage{
			{Role: "user", Author: "alice", Platform: "slack", Content: "hi", Timestamp: now},
			{Role: "assistant", Author: "agent", Platform: "slack", Content: "hello", Timestamp: now},
		}

		initSuite := func(suite *WhenWritingConversationHistoryTestingSuite) {
			suite.mockAppendService.EXPECT().
				Append(conversationId, gomock.Any()).
				DoAndReturn(func(_ string, msgs []models.ConversationMessage) ([]models.ConversationMessage, error) {
					if assert.Len(t, msgs, 2) {
						assert.Equal(t, "user", msgs[0].Role)
						assert.Equal(t, "alice", msgs[0].Author)
						assert.Equal(t, "slack", msgs[0].Platform)
						assert.Equal(t, "hi", msgs[0].Content)
						assert.Equal(t, now, msgs[0].Timestamp)
						assert.Equal(t, "assistant", msgs[1].Role)
					}
					return msgs, nil
				})
		}

		t.Run("Should map DTOs to persistence models and delegate to the append service", func(t *testing.T) {
			t.Parallel()

			suite := WhenWritingConversationHistoryBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Write(conversationId, inputs)
			assert.NoError(t, err)
			assert.Len(t, result, 2)
		})
	})

	t.Run("Given the append service returns persistence rows", func(t *testing.T) {
		t.Parallel()

		persisted := []models.ConversationMessage{
			{
				Role:           "user",
				Author:         "alice",
				Platform:       "slack",
				Content:        "hi",
				ConversationId: conversationId,
				Hash:           "computed-by-service",
				Timestamp:      now,
			},
		}

		initSuite := func(suite *WhenWritingConversationHistoryTestingSuite) {
			suite.mockAppendService.EXPECT().
				Append(conversationId, gomock.Any()).
				Return(persisted, nil)
		}

		t.Run("Should map them back to DTOs (no Hash, no ConversationId leak)", func(t *testing.T) {
			t.Parallel()

			suite := WhenWritingConversationHistoryBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Write(conversationId, []interfaces.ConversationMessage{
				{Role: "user", Author: "alice", Platform: "slack", Content: "hi", Timestamp: now},
			})
			assert.NoError(t, err)
			if assert.Len(t, result, 1) {
				assert.Equal(t, "alice", result[0].Author)
				assert.Equal(t, "hi", result[0].Content)
				assert.Equal(t, now, result[0].Timestamp)
			}
		})
	})

	t.Run("Given the append service fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenWritingConversationHistoryTestingSuite) {
			suite.mockAppendService.EXPECT().
				Append(conversationId, gomock.Any()).
				Return(nil, errors.New("db down"))
		}

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenWritingConversationHistoryBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Write(conversationId, []interfaces.ConversationMessage{
				{Role: "user", Content: "hi"},
			})
			assert.ErrorContains(t, err, "db down")
		})
	})
}
