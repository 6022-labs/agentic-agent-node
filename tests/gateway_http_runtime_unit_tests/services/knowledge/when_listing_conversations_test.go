package services_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	services "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services/knowledge"
	clients_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_http_runtime_mocks/clients_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenListingConversationsTestingSuite struct {
	sut *services.HttpConversationHistoryClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenListingConversationsBeforeEach(t *testing.T) *WhenListingConversationsTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenListingConversationsTestingSuite{
		sut:        services.NewHttpConversationHistoryClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenListingConversations(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime returns the conversations", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingConversationsTestingSuite) {
			suite.mockClient.EXPECT().GetInternalConversationsWithResponse(gomock.Any()).Return(&clients.GetInternalConversationsResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusOK},
				Body:         []byte(`{"conversations":[{"conversationId":"a","messageCount":2,"firstMessageAt":"2026-01-01T00:00:00Z","lastMessageAt":"2026-01-02T00:00:00Z"}]}`),
			}, nil)
		}

		t.Run("Should decode the conversation summary list payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationsBeforeEach(t)
			initSuite(suite)

			payload, err := suite.sut.ListConversations(context.Background())

			assert.NoError(t, err)
			if assert.NotNil(t, payload) && assert.Len(t, payload.Conversations, 1) {
				assert.Equal(t, "a", payload.Conversations[0].ConversationId)
				assert.Equal(t, int64(2), payload.Conversations[0].MessageCount)
			}
		})
	})

	t.Run("Given listing fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingConversationsTestingSuite) {
			suite.mockClient.EXPECT().GetInternalConversationsWithResponse(gomock.Any()).Return(nil, errors.New("runtime down"))
		}

		t.Run("Should propagate the transport error", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationsBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.ListConversations(context.Background())

			assert.ErrorContains(t, err, "runtime down")
		})
	})
}
