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
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenListingConversationsTestingSuite struct {
	sut *services.HttpConversationClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenListingConversationsBeforeEach(t *testing.T) *WhenListingConversationsTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenListingConversationsTestingSuite{
		sut:        services.NewHttpConversationClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenListingConversations(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingConversationsTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalConversationsWithResponse(gomock.Any(), gomock.Any()).
				Return(nil, errors.New("runtime down"))
		}

		t.Run("Should map the transport error onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationsBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.ListConversations(context.Background(), 1, 10)

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Equal(t, "runtime_unavailable", unavailableErr.Code)
			assert.Contains(t, unavailableErr.Message, "runtime down")
		})
	})

	t.Run("Given the runtime answers 400", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingConversationsTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalConversationsWithResponse(gomock.Any(), gomock.Any()).
				Return(&clients.GetInternalConversationsResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusBadRequest},
					Body:         []byte(`{"field":"callerId","message":"too long"}`),
				}, nil)
		}

		t.Run("Should map the 400 onto a ValidationError", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationsBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.ListConversations(context.Background(), 1, 10)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "callerId", validationErr.Field)
		})
	})

	t.Run("Given the runtime receives no caller id", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingConversationsTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalConversationsWithResponse(gomock.Any(), gomock.Any()).
				Return(&clients.GetInternalConversationsResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusUnauthorized},
					Body: []byte(
						`{"code":"missing_caller_id","message":"an authenticated caller id is required"}`,
					),
				}, nil)
		}

		t.Run("Should map the 401 onto an UnauthorizedError", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationsBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.ListConversations(context.Background(), 1, 10)

			var unauthorizedErr *apperrors.UnauthorizedError
			assert.ErrorAs(t, err, &unauthorizedErr)
			assert.Equal(t, "missing_caller_id", unauthorizedErr.Code)
		})
	})

	t.Run("Given the runtime answers 500", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingConversationsTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalConversationsWithResponse(gomock.Any(), gomock.Any()).
				Return(&clients.GetInternalConversationsResponse{HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError}, Body: []byte(`{"code":"x","field":"f","message":"m"}`)}, nil)
		}

		t.Run("Should map the 500 onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationsBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.ListConversations(context.Background(), 1, 10)

			var target *apperrors.UnavailableError
			if assert.ErrorAs(t, err, &target) {
				assert.Equal(t, "runtime_unavailable", target.Code)
			}
		})
	})

	t.Run("Given the runtime returns the conversations", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingConversationsTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalConversationsWithResponse(gomock.Any(), gomock.Any()).
				DoAndReturn(func(
					_ context.Context,
					params *clients.GetInternalConversationsParams,
					_ ...clients.RequestEditorFn,
				) (*clients.GetInternalConversationsResponse, error) {
					assert.Equal(t, 2, *params.Page)
					assert.Equal(t, 25, *params.ElementsPerPage)
					return &clients.GetInternalConversationsResponse{
						HTTPResponse: &http.Response{StatusCode: http.StatusOK},
						Body: []byte(
							`{"page":2,"elementsPerPage":25,"totalElements":1,` +
								`"conversations":[{"conversationId":"a","name":"ERC20 mint audit","messageCount":2,` +
								`"firstMessageAt":"2026-01-01T00:00:00Z","lastMessageAt":"2026-01-02T00:00:00Z",` +
								`"createdAt":"2026-01-01T00:00:00Z"}]}`,
						),
					}, nil
				})
		}

		t.Run("Should decode the conversation summary list payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingConversationsBeforeEach(t)
			initSuite(suite)

			payload, err := suite.sut.ListConversations(context.Background(), 2, 25)

			assert.NoError(t, err)
			if assert.NotNil(t, payload) && assert.Len(t, payload.Conversations, 1) {
				assert.Equal(t, uint(2), payload.Page)
				assert.Equal(t, uint(1), payload.TotalElements)
				assert.Equal(t, "a", payload.Conversations[0].ConversationId)
				assert.Equal(t, "ERC20 mint audit", *payload.Conversations[0].Name)
				assert.Equal(t, int64(2), payload.Conversations[0].MessageCount)
			}
		})
	})
}
