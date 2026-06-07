package use_cases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenListingConversationMessagesTestingSuite struct {
	sut *use_cases.ListConversationMessages

	mockClient *interfaces_mocks.MockConversationHistoryClientInterface
}

func WhenListingConversationMessagesBeforeEach(t *testing.T) *WhenListingConversationMessagesTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := interfaces_mocks.NewMockConversationHistoryClientInterface(mockController)

	return &WhenListingConversationMessagesTestingSuite{
		sut:        use_cases.NewListConversationMessages(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenListingConversationMessagesViaGateway(t *testing.T) {
	t.Parallel()

	conversationId := "conv-1"

	t.Run("Given the client returns messages", func(t *testing.T) {
		t.Parallel()

		now := time.Now().UTC().Truncate(time.Microsecond)

		initSuite := func(suite *WhenListingConversationMessagesTestingSuite) {
			suite.mockClient.EXPECT().ListMessages(gomock.Any(), conversationId).Return(&models.ExternalListConversationMessagesPayload{
				Messages: []models.ExternalConversationMessage{
					{Id: "m1", ConversationId: conversationId, Role: "user", Content: "hi", Timestamp: now},
				},
			}, nil)
		}

		t.Run("Should map them onto the wire response", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationMessagesBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background(), conversationId)

			assert.NoError(t, err)
			assert.Len(t, resp.Messages, 1)
			assert.Equal(t, "m1", resp.Messages[0].Id)
			assert.Equal(t, "hi", resp.Messages[0].Content)
			assert.Equal(t, "user", resp.Messages[0].Role)
		})
	})

	t.Run("Given the client fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingConversationMessagesTestingSuite) {
			suite.mockClient.EXPECT().ListMessages(gomock.Any(), conversationId).Return(nil, errors.New("runtime down"))
		}

		t.Run("Should return Unavailable carrying the message", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), conversationId)

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Contains(t, unavailableErr.Message, "runtime down")
		})
	})
}
