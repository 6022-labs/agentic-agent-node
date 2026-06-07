package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/services"
	"github.com/6022-labs/agentic-agent-node/tests/agent_knowledge_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenReadingConversationHistoryTestingSuite struct {
	sut *services.ConversationHistory

	mockListConversationMessages   *interfaces_mocks.MockListConversationMessagesInterface
	mockAppendConversationMessages *interfaces_mocks.MockAppendConversationMessagesInterface
}

func WhenReadingConversationHistoryBeforeEach(t *testing.T) *WhenReadingConversationHistoryTestingSuite {
	mockController := gomock.NewController(t)

	mockListConversationMessages := interfaces_mocks.NewMockListConversationMessagesInterface(mockController)
	mockAppendConversationMessages := interfaces_mocks.NewMockAppendConversationMessagesInterface(mockController)

	sut := services.NewConversationHistory(mockListConversationMessages, mockAppendConversationMessages)

	return &WhenReadingConversationHistoryTestingSuite{
		sut:                            sut,
		mockListConversationMessages:   mockListConversationMessages,
		mockAppendConversationMessages: mockAppendConversationMessages,
	}
}

func TestWhenReadingConversationHistory(t *testing.T) {
	t.Parallel()

	conversationId := "conv-1"
	now := time.Now().UTC().Truncate(time.Microsecond)

	t.Run("Given the list service returns persistence rows", func(t *testing.T) {
		t.Parallel()

		stored := []models.ConversationMessage{
			{
				Role:           "user",
				Content:        "hi :fire:",
				ConversationId: conversationId,
				Hash:           "ignored-by-dto",
				Timestamp:      now,
			},
		}

		initSuite := func(suite *WhenReadingConversationHistoryTestingSuite) {
			suite.mockListConversationMessages.EXPECT().List(gomock.Any(), conversationId).Return(stored, nil)
		}

		t.Run("Should return DTOs with only the semantic fields exposed", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingConversationHistoryBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Read(context.Background(), conversationId)
			assert.NoError(t, err)
			if assert.Len(t, result, 1) {
				assert.Equal(t, "user", result[0].Role)
				assert.Equal(t, "hi :fire:", result[0].Content)
				assert.Equal(t, now, result[0].Timestamp)
			}
		})
	})

	t.Run("Given the list service returns an empty slice", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenReadingConversationHistoryTestingSuite) {
			suite.mockListConversationMessages.EXPECT().List(gomock.Any(), conversationId).Return([]models.ConversationMessage{}, nil)
		}

		t.Run("Should return an empty DTO slice", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingConversationHistoryBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Read(context.Background(), conversationId)
			assert.NoError(t, err)
			assert.Empty(t, result)
		})
	})

	t.Run("Given the list service fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenReadingConversationHistoryTestingSuite) {
			suite.mockListConversationMessages.EXPECT().List(gomock.Any(), conversationId).Return(nil, errors.New("db down"))
		}

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingConversationHistoryBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Read(context.Background(), conversationId)
			assert.ErrorContains(t, err, "db down")
		})
	})
}
