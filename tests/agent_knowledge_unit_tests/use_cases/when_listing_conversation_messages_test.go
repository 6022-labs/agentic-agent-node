package use_cases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/tests/agent_knowledge_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenListingConversationMessagesTestingSuite struct {
	sut *use_cases.ListConversationMessages

	mockListConversationMessages *interfaces_mocks.MockListConversationMessagesInterface
}

func WhenListingConversationMessagesBeforeEach(t *testing.T) *WhenListingConversationMessagesTestingSuite {
	mockController := gomock.NewController(t)

	mockListConversationMessages := interfaces_mocks.NewMockListConversationMessagesInterface(mockController)

	sut := use_cases.NewListConversationMessages(mockListConversationMessages)

	return &WhenListingConversationMessagesTestingSuite{
		sut:                          sut,
		mockListConversationMessages: mockListConversationMessages,
	}
}

func TestWhenListingConversationMessages(t *testing.T) {
	t.Parallel()

	conversationId := "conv-1"

	t.Run("Given the service returns rows", func(t *testing.T) {
		t.Parallel()

		now := time.Now()
		stored := []models.ConversationMessage{
			model_builders.NewConversationMessageBuilder().WithConversationId(conversationId).WithTimestamp(now).Build(),
		}

		initSuite := func(suite *WhenListingConversationMessagesTestingSuite) {
			suite.mockListConversationMessages.EXPECT().List(gomock.Any(), conversationId).Return(stored, nil)
		}

		t.Run("Should return Ok with the rows on the wire", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationMessagesBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background(), conversationId)

			assert.NoError(t, err)
			assert.Len(t, resp.Messages, 1)
			assert.Equal(t, conversationId, resp.Messages[0].ConversationId)
		})
	})

	t.Run("Given the service fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingConversationMessagesTestingSuite) {
			suite.mockListConversationMessages.EXPECT().List(gomock.Any(), conversationId).Return(nil, errors.New("db blew"))
		}

		t.Run("Should return InternalError carrying the service message", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), conversationId)

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Contains(t, internalErr.Message, "db blew")
		})
	})
}
