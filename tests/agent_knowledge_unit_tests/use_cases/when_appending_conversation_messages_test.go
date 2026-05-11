package use_cases_test

import (
	"errors"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases"
	"github.com/6022-labs/agentic-agent-node/tests/agent_knowledge_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
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

		input := []models.ConversationMessage{
			model_builders.NewConversationMessageBuilder().WithRole("user").WithContent("hi").WithTimestamp(now).Build(),
		}

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			suite.mockService.EXPECT().Append(conversationId, input).Return(input, nil)
		}

		t.Run("Should forward conversationId + messages and return the persisted rows", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			res, err := suite.sut.Execute(conversationId, input)

			assert.NoError(t, err)
			assert.Equal(t, input, res)
		})
	})

	t.Run("Given the service fails", func(t *testing.T) {
		t.Parallel()

		input := []models.ConversationMessage{
			model_builders.NewConversationMessageBuilder().WithTimestamp(now).Build(),
		}

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			suite.mockService.EXPECT().Append(conversationId, input).Return(nil, errors.New("db down"))
		}

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(conversationId, input)

			assert.ErrorContains(t, err, "db down")
		})
	})
}
