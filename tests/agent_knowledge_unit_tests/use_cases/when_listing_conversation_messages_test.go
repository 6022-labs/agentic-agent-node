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

type WhenListingConversationMessagesTestingSuite struct {
	sut *use_cases.ListConversationMessages

	mockService *interfaces_mocks.MockListConversationMessagesInterface
}

func WhenListingConversationMessagesBeforeEach(t *testing.T) *WhenListingConversationMessagesTestingSuite {
	mockController := gomock.NewController(t)

	mockService := interfaces_mocks.NewMockListConversationMessagesInterface(mockController)

	sut := use_cases.NewListConversationMessages(mockService)

	return &WhenListingConversationMessagesTestingSuite{
		sut:         sut,
		mockService: mockService,
	}
}

func TestWhenListingConversationMessages(t *testing.T) {
	t.Parallel()

	conversationId := "conv-1"

	t.Run("Given the service returns rows", func(t *testing.T) {
		t.Parallel()

		now := time.Now()
		expected := []models.ConversationMessage{
			model_builders.NewConversationMessageBuilder().WithConversationId(conversationId).WithTimestamp(now).Build(),
		}

		initSuite := func(suite *WhenListingConversationMessagesTestingSuite) {
			suite.mockService.EXPECT().List(conversationId).Return(expected, nil)
		}

		t.Run("Should forward the conversationId and return the rows verbatim", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationMessagesBeforeEach(t)
			initSuite(suite)

			res, err := suite.sut.Execute(conversationId)

			assert.NoError(t, err)
			assert.Equal(t, expected, res)
		})
	})

	t.Run("Given the service fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingConversationMessagesTestingSuite) {
			suite.mockService.EXPECT().List(conversationId).Return(nil, errors.New("db blew"))
		}

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(conversationId)

			assert.ErrorContains(t, err, "db blew")
		})
	})
}
