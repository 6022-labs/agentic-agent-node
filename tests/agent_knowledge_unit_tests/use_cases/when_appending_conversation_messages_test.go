package use_cases_test

import (
	"errors"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/usecase_result"
	"github.com/6022-labs/agentic-agent-node/tests/agent_knowledge_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenAppendingConversationMessagesTestingSuite struct {
	sut *use_cases.AppendConversationMessages

	mockService *interfaces_mocks.MockAppendConversationMessagesInterface
}

func WhenAppendingConversationMessagesBeforeEach(t *testing.T) *WhenAppendingConversationMessagesTestingSuite {
	mockController := gomock.NewController(t)

	mockService := interfaces_mocks.NewMockAppendConversationMessagesInterface(mockController)

	sut := use_cases.NewAppendConversationMessages(mockService)

	return &WhenAppendingConversationMessagesTestingSuite{
		sut:         sut,
		mockService: mockService,
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
			suite.mockService.EXPECT().Append(conversationId, mapped).Return(mapped, nil)
		}

		t.Run("Should return Ok with the persisted rows on the wire", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			res := suite.sut.Execute(conversationId, req)

			assert.Equal(t, usecase_result.OutcomeOk, res.Outcome)
			body, ok := res.Body.(*responses.AppendConversationMessagesResponse)
			assert.True(t, ok)
			assert.Len(t, body.Inserted, 1)
			assert.Equal(t, "user", body.Inserted[0].Role)
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
			suite.mockService.EXPECT().Append(conversationId, mapped).Return(nil, errors.New("db down"))
		}

		t.Run("Should return InternalError carrying the service message", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			res := suite.sut.Execute(conversationId, req)

			assert.Equal(t, usecase_result.OutcomeInternalError, res.Outcome)
			assert.Contains(t, res.Message, "db down")
		})
	})
}
