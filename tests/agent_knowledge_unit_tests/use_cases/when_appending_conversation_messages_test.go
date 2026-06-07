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
	"github.com/6022-labs/agentic-agent-node/tests/agent_knowledge_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenAppendingConversationMessagesTestingSuite struct {
	sut *use_cases.AppendConversationMessages

	mockAppendConversationMessages *interfaces_mocks.MockAppendConversationMessagesInterface
}

func WhenAppendingConversationMessagesBeforeEach(t *testing.T) *WhenAppendingConversationMessagesTestingSuite {
	mockController := gomock.NewController(t)

	mockAppendConversationMessages := interfaces_mocks.NewMockAppendConversationMessagesInterface(mockController)

	sut := use_cases.NewAppendConversationMessages(mockAppendConversationMessages)

	return &WhenAppendingConversationMessagesTestingSuite{
		sut:                            sut,
		mockAppendConversationMessages: mockAppendConversationMessages,
	}
}

func TestWhenAppendingConversationMessages(t *testing.T) {
	t.Parallel()

	now := time.Now()
	conversationId := "conv-1"

	t.Run("Given the service persists the messages", func(t *testing.T) {
		t.Parallel()

		req := &requests.AppendConversationMessagesRequest{
			Messages: []requests.ConversationMessageInput{
				{Role: "user", Content: "hi", Timestamp: now},
			},
		}
		mapped := []models.ConversationMessage{
			{Role: "user", Content: "hi", Timestamp: now},
		}

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			suite.mockAppendConversationMessages.EXPECT().Append(gomock.Any(), conversationId, mapped).Return(mapped, nil)
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

	t.Run("Given the service fails", func(t *testing.T) {
		t.Parallel()

		req := &requests.AppendConversationMessagesRequest{
			Messages: []requests.ConversationMessageInput{
				{Timestamp: now},
			},
		}
		mapped := []models.ConversationMessage{
			{Timestamp: now},
		}

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			suite.mockAppendConversationMessages.EXPECT().Append(gomock.Any(), conversationId, mapped).Return(nil, errors.New("db down"))
		}

		t.Run("Should return InternalError carrying the service message", func(t *testing.T) {
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
