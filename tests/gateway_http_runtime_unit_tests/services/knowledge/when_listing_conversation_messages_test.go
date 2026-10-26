package services_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
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

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingConversationMessagesTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalConversationsConversationIdMessagesWithResponse(gomock.Any(), conversationId).
				Return(nil, errors.New("runtime down"))
		}

		t.Run("Should map the transport error onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.ListMessages(context.Background(), conversationId)

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Equal(t, "runtime_unavailable", unavailableErr.Code)
			assert.NotContains(t, unavailableErr.Message, "runtime down")
		})
	})

	t.Run("Given the runtime answers 400", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingConversationMessagesTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalConversationsConversationIdMessagesWithResponse(gomock.Any(), conversationId).
				Return(&clients.GetInternalConversationsConversationIdMessagesResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusBadRequest},
					Body:         []byte(`{"field":"conversationId","message":"cannot be empty"}`),
				}, nil)
		}

		t.Run("Should map the 400 onto a ValidationError", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.ListMessages(context.Background(), conversationId)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "conversationId", validationErr.Field)
		})
	})

	t.Run("Given the runtime receives no caller id", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingConversationMessagesTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalConversationsConversationIdMessagesWithResponse(gomock.Any(), conversationId).
				Return(&clients.GetInternalConversationsConversationIdMessagesResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusUnauthorized},
					Body: []byte(
						`{"code":"missing_caller_id","message":"an authenticated caller id is required"}`,
					),
				}, nil)
		}

		t.Run("Should map the 401 onto an UnauthorizedError", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.ListMessages(context.Background(), conversationId)

			var unauthorizedErr *apperrors.UnauthorizedError
			assert.ErrorAs(t, err, &unauthorizedErr)
			assert.Equal(t, "missing_caller_id", unauthorizedErr.Code)
		})
	})

	t.Run("Given the conversation is out of the caller's view", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingConversationMessagesTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalConversationsConversationIdMessagesWithResponse(gomock.Any(), conversationId).
				Return(&clients.GetInternalConversationsConversationIdMessagesResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusNotFound},
					Body:         []byte(`{"code":"conversation_not_found","message":"conversation not found"}`),
				}, nil)
		}

		t.Run("Should map the 404 onto a NotFoundError", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.ListMessages(context.Background(), conversationId)

			var notFoundErr *apperrors.NotFoundError
			assert.ErrorAs(t, err, &notFoundErr)
			assert.Equal(t, "conversation_not_found", notFoundErr.Code)
		})
	})

	t.Run("Given the runtime answers 500", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingConversationMessagesTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalConversationsConversationIdMessagesWithResponse(gomock.Any(), gomock.Any()).
				Return(&clients.GetInternalConversationsConversationIdMessagesResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError},
					Body:         []byte(`{"code":"x","field":"f","message":"m"}`),
				}, nil)
		}

		t.Run("Should map the 500 onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.ListMessages(context.Background(), conversationId)

			var target *apperrors.UnavailableError
			if assert.ErrorAs(t, err, &target) {
				assert.Equal(t, "runtime_unavailable", target.Code)
			}
		})
	})

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
}
