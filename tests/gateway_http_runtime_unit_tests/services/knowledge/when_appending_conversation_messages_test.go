package services_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	services "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services/knowledge"
	clients_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_http_runtime_mocks/clients_mocks"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenAppendingConversationMessagesTestingSuite struct {
	sut *services.HttpConversationClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenAppendingConversationMessagesBeforeEach(t *testing.T) *WhenAppendingConversationMessagesTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenAppendingConversationMessagesTestingSuite{
		sut:        services.NewHttpConversationClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenAppendingConversationMessages(t *testing.T) {
	t.Parallel()

	conversationId := "conv-1"
	now := time.Now().UTC().Truncate(time.Microsecond)

	request := &models.ExternalAppendConversationMessagesPayload{
		Messages: []models.ExternalConversationMessageInput{
			{Role: "user", Content: "fake memory", Timestamp: now},
		},
	}

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalConversationsConversationIdMessagesWithBodyWithResponse(
					gomock.Any(),
					conversationId,
					"application/json",
					gomock.Any(),
				).
				Return(nil, errors.New("runtime down"))
		}

		t.Run("Should map the transport error onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.AppendMessages(context.Background(), conversationId, request)

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Equal(t, "runtime_unavailable", unavailableErr.Code)
			assert.Contains(t, unavailableErr.Message, "runtime down")
		})
	})

	t.Run("Given the runtime answers 400", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalConversationsConversationIdMessagesWithBodyWithResponse(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).
				Return(&clients.PostInternalConversationsConversationIdMessagesResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusBadRequest},
					Body:         []byte(`{"code":"x","field":"f","message":"m"}`),
				}, nil)
		}

		t.Run("Should map the 400 onto a ValidationError", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.AppendMessages(context.Background(), conversationId, request)

			var target *apperrors.ValidationError
			assert.ErrorAs(t, err, &target)
		})
	})

	t.Run("Given the runtime receives no caller id", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalConversationsConversationIdMessagesWithBodyWithResponse(
					gomock.Any(),
					conversationId,
					"application/json",
					gomock.Any(),
				).
				Return(&clients.PostInternalConversationsConversationIdMessagesResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusUnauthorized},
					Body: []byte(
						`{"code":"missing_caller_id","message":"an authenticated caller id is required"}`,
					),
				}, nil)
		}

		t.Run("Should map the 401 onto an UnauthorizedError", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.AppendMessages(context.Background(), conversationId, request)

			var unauthorizedErr *apperrors.UnauthorizedError
			assert.ErrorAs(t, err, &unauthorizedErr)
			assert.Equal(t, "missing_caller_id", unauthorizedErr.Code)
		})
	})

	t.Run("Given the conversation is out of the caller's view", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalConversationsConversationIdMessagesWithBodyWithResponse(
					gomock.Any(),
					conversationId,
					"application/json",
					gomock.Any(),
				).
				Return(&clients.PostInternalConversationsConversationIdMessagesResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusNotFound},
					Body:         []byte(`{"code":"conversation_not_found","message":"conversation not found"}`),
				}, nil)
		}

		t.Run("Should map the 404 onto a NotFoundError", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.AppendMessages(context.Background(), conversationId, request)

			var notFoundErr *apperrors.NotFoundError
			assert.ErrorAs(t, err, &notFoundErr)
			assert.Equal(t, "conversation_not_found", notFoundErr.Code)
		})
	})

	t.Run("Given the runtime refuses because the agent is not minted", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalConversationsConversationIdMessagesWithBodyWithResponse(
					gomock.Any(),
					conversationId,
					"application/json",
					gomock.Any(),
				).
				Return(&clients.PostInternalConversationsConversationIdMessagesResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusConflict},
					Body:         []byte(`{"code":"agent_not_minted","message":"agent identity is not minted yet"}`),
				}, nil)
		}

		t.Run("Should map the 409 onto a ConflictError", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.AppendMessages(context.Background(), conversationId, request)

			var conflictErr *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictErr)
			assert.Equal(t, "agent_not_minted", conflictErr.Code)
		})
	})

	t.Run("Given the runtime answers 500", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalConversationsConversationIdMessagesWithBodyWithResponse(
					gomock.Any(),
					conversationId,
					"application/json",
					gomock.Any(),
				).
				Return(&clients.PostInternalConversationsConversationIdMessagesResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError},
					Body:         []byte(`{"code":"boom","message":"runtime down"}`),
				}, nil)
		}

		t.Run("Should map the 500 onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.AppendMessages(context.Background(), conversationId, request)

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Equal(t, "runtime_unavailable", unavailableErr.Code)
		})
	})

	t.Run("Given the runtime accepts the messages", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalConversationsConversationIdMessagesWithBodyWithResponse(
					gomock.Any(),
					conversationId,
					"application/json",
					gomock.Any(),
				).
				DoAndReturn(
					func(_ context.Context, _ string, _ string, reader io.Reader, _ ...clients.RequestEditorFn) (*clients.PostInternalConversationsConversationIdMessagesResponse, error) {
						raw, err := io.ReadAll(reader)
						assert.NoError(t, err)

						var sent models.ExternalAppendConversationMessagesPayload
						assert.NoError(t, json.Unmarshal(raw, &sent))
						assert.Len(t, sent.Messages, 1)
						assert.Equal(t, "fake memory", sent.Messages[0].Content)

						return &clients.PostInternalConversationsConversationIdMessagesResponse{
							HTTPResponse: &http.Response{StatusCode: http.StatusOK},
							Body: []byte(
								`{"inserted":[{"id":"m1","conversationId":"conv-1","role":"user",` +
									`"content":"fake memory","timestamp":"2026-01-01T00:00:00Z"}]}`,
							),
						}, nil
					})
		}

		t.Run("Should decode the inserted messages payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			payload, err := suite.sut.AppendMessages(context.Background(), conversationId, request)

			assert.NoError(t, err)
			if assert.NotNil(t, payload) && assert.Len(t, payload.Inserted, 1) {
				assert.Equal(t, "m1", payload.Inserted[0].Id)
				assert.Equal(t, "fake memory", lo.FromPtr(payload.Inserted[0].Content))
			}
		})
	})
}
