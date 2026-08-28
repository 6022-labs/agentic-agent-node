package services_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	services "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services/knowledge"
	clients_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_http_runtime_mocks/clients_mocks"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenListingConversationMessagesTestingSuite struct {
	sut *services.HttpConversationClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenListingConversationMessagesBeforeEach(t *testing.T) *WhenListingConversationMessagesTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenListingConversationMessagesTestingSuite{
		sut:        services.NewHttpConversationClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenListingConversationMessages(t *testing.T) {
	t.Parallel()

	conversationId := "conv-1"

	t.Run("Given the runtime returns the messages", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingConversationMessagesTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalConversationsConversationIdMessagesWithResponse(gomock.Any(), conversationId).
				Return(&clients.GetInternalConversationsConversationIdMessagesResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusOK},
					Body: []byte(
						`{"messages":[{"id":"m1","conversationId":"conv-1","role":"user",` +
							`"content":"hi","timestamp":"2026-01-01T00:00:00Z"}]}`,
					),
				}, nil)
		}

		t.Run("Should decode the conversation messages payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationMessagesBeforeEach(t)
			initSuite(suite)

			payload, err := suite.sut.ListMessages(context.Background(), conversationId)

			assert.NoError(t, err)
			if assert.NotNil(t, payload) && assert.Len(t, payload.Messages, 1) {
				assert.Equal(t, "m1", payload.Messages[0].Id)
				assert.Equal(t, "hi", lo.FromPtr(payload.Messages[0].Content))
			}
		})
	})

	t.Run("Given listing fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingConversationMessagesTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalConversationsConversationIdMessagesWithResponse(gomock.Any(), conversationId).
				Return(nil, errors.New("runtime down"))
		}

		t.Run("Should propagate the transport error", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.ListMessages(context.Background(), conversationId)

			assert.ErrorContains(t, err, "runtime down")
		})
	})
}
