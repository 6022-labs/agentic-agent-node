package services_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	services "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services/knowledge"
	clients_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_http_runtime_mocks/clients_mocks"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenUpdatingConversationTestingSuite struct {
	sut *services.HttpConversationClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenUpdatingConversationBeforeEach(t *testing.T) *WhenUpdatingConversationTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenUpdatingConversationTestingSuite{
		sut:        services.NewHttpConversationClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenUpdatingConversation(t *testing.T) {
	t.Parallel()

	conversationId := "c1"
	request := &models.ExternalUpdateConversationPayload{Name: "Solidity audits"}

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingConversationTestingSuite) {
			suite.mockClient.EXPECT().
				PutInternalConversationsConversationIdWithBodyWithResponse(
					gomock.Any(),
					conversationId,
					"application/json",
					gomock.Any(),
				).
				Return(nil, errors.New("runtime down"))
		}

		t.Run("Should map the transport error onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingConversationBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.UpdateConversation(context.Background(), conversationId, request)

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Equal(t, "runtime_unavailable", unavailableErr.Code)
			assert.Contains(t, unavailableErr.Message, "runtime down")
		})
	})

	t.Run("Given the runtime answers 400", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingConversationTestingSuite) {
			suite.mockClient.EXPECT().
				PutInternalConversationsConversationIdWithBodyWithResponse(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&clients.PutInternalConversationsConversationIdResponse{HTTPResponse: &http.Response{StatusCode: http.StatusBadRequest}, Body: []byte(`{"code":"x","field":"f","message":"m"}`)}, nil)
		}

		t.Run("Should map the 400 onto a ValidationError", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingConversationBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.UpdateConversation(context.Background(), conversationId, request)

			var target *apperrors.ValidationError
			assert.ErrorAs(t, err, &target)
		})
	})

	t.Run("Given the runtime receives no caller id", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingConversationTestingSuite) {
			suite.mockClient.EXPECT().
				PutInternalConversationsConversationIdWithBodyWithResponse(
					gomock.Any(),
					conversationId,
					"application/json",
					gomock.Any(),
				).
				Return(&clients.PutInternalConversationsConversationIdResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusUnauthorized},
					Body: []byte(
						`{"code":"missing_caller_id","message":"an authenticated caller id is required"}`,
					),
				}, nil)
		}

		t.Run("Should map the 401 onto an UnauthorizedError", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingConversationBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.UpdateConversation(context.Background(), conversationId, request)

			var unauthorizedErr *apperrors.UnauthorizedError
			assert.ErrorAs(t, err, &unauthorizedErr)
			assert.Equal(t, "missing_caller_id", unauthorizedErr.Code)
		})
	})

	t.Run("Given the conversation is out of the caller's view", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingConversationTestingSuite) {
			suite.mockClient.EXPECT().
				PutInternalConversationsConversationIdWithBodyWithResponse(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).
				Return(&clients.PutInternalConversationsConversationIdResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusNotFound},
					Body:         []byte(`{"code":"conversation_not_found","message":"conversation not found"}`),
				}, nil)
		}

		t.Run("Should map the 404 onto a NotFoundError", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingConversationBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.UpdateConversation(context.Background(), "ghost", request)

			var notFoundErr *apperrors.NotFoundError
			assert.ErrorAs(t, err, &notFoundErr)
			assert.Equal(t, "conversation_not_found", notFoundErr.Code)
		})
	})

	t.Run("Given another conversation holds that name", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingConversationTestingSuite) {
			suite.mockClient.EXPECT().
				PutInternalConversationsConversationIdWithBodyWithResponse(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&clients.PutInternalConversationsConversationIdResponse{HTTPResponse: &http.Response{StatusCode: http.StatusConflict}, Body: []byte(`{"code":"conversation_name_taken","message":"already used"}`)}, nil)
		}

		t.Run("Should map the 409 onto a ConflictError", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingConversationBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.UpdateConversation(context.Background(), conversationId, request)

			var target *apperrors.ConflictError
			assert.ErrorAs(t, err, &target)
		})
	})

	t.Run("Given the runtime answers 500", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingConversationTestingSuite) {
			suite.mockClient.EXPECT().
				PutInternalConversationsConversationIdWithBodyWithResponse(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&clients.PutInternalConversationsConversationIdResponse{HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError}, Body: []byte(`{"code":"x","field":"f","message":"m"}`)}, nil)
		}

		t.Run("Should map the 500 onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingConversationBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.UpdateConversation(context.Background(), conversationId, request)

			var target *apperrors.UnavailableError
			if assert.ErrorAs(t, err, &target) {
				assert.Equal(t, "runtime_unavailable", target.Code)
			}
		})
	})

	t.Run("Given the runtime renames the session", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenUpdatingConversationTestingSuite) {
			suite.mockClient.EXPECT().
				PutInternalConversationsConversationIdWithBodyWithResponse(
					gomock.Any(),
					conversationId,
					"application/json",
					gomock.Any(),
				).
				DoAndReturn(
					func(_ context.Context, _ string, _ string, reader io.Reader, _ ...clients.RequestEditorFn) (*clients.PutInternalConversationsConversationIdResponse, error) {
						raw, err := io.ReadAll(reader)
						assert.NoError(t, err)

						var sent models.ExternalUpdateConversationPayload
						assert.NoError(t, json.Unmarshal(raw, &sent))
						assert.Equal(t, "Solidity audits", sent.Name)

						return &clients.PutInternalConversationsConversationIdResponse{
							HTTPResponse: &http.Response{StatusCode: http.StatusOK},
							Body: []byte(
								`{"conversationId":"c1","name":"Solidity audits","messageCount":2,` +
									`"createdAt":"2026-01-01T00:00:00Z","updatedAt":"2026-01-02T00:00:00Z"}`,
							),
						}, nil
					})
		}

		t.Run("Should decode the renamed conversation", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingConversationBeforeEach(t)
			initSuite(suite)

			payload, err := suite.sut.UpdateConversation(context.Background(), conversationId, request)

			assert.NoError(t, err)
			if assert.NotNil(t, payload) {
				assert.Equal(t, "Solidity audits", lo.FromPtr(payload.Name))
				assert.Equal(t, int64(2), payload.MessageCount)
			}
		})
	})
}
