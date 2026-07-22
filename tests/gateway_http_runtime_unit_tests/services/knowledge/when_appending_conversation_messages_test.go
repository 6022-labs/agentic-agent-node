package services_test

import (
	"context"
	"encoding/json"
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
	sut *services.HttpConversationHistoryClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenAppendingConversationMessagesBeforeEach(t *testing.T) *WhenAppendingConversationMessagesTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenAppendingConversationMessagesTestingSuite{
		sut:        services.NewHttpConversationHistoryClient(mockClient),
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

	t.Run("Given the runtime accepts the messages", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			suite.mockClient.EXPECT().PostInternalConversationsConversationIdMessagesWithBodyWithResponse(gomock.Any(), conversationId, "application/json", gomock.Any()).DoAndReturn(
				func(_ context.Context, _ string, _ string, reader io.Reader, _ ...clients.RequestEditorFn) (*clients.PostInternalConversationsConversationIdMessagesResponse, error) {
					raw, err := io.ReadAll(reader)
					assert.NoError(t, err)

					var sent models.ExternalAppendConversationMessagesPayload
					assert.NoError(t, json.Unmarshal(raw, &sent))
					assert.Len(t, sent.Messages, 1)
					assert.Equal(t, "fake memory", sent.Messages[0].Content)

					return &clients.PostInternalConversationsConversationIdMessagesResponse{
						HTTPResponse: &http.Response{StatusCode: http.StatusOK},
						Body:         []byte(`{"inserted":[{"id":"m1","conversationId":"conv-1","role":"user","content":"fake memory","timestamp":"2026-01-01T00:00:00Z"}]}`),
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

	t.Run("Given the runtime refuses because the agent is not minted", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			suite.mockClient.EXPECT().PostInternalConversationsConversationIdMessagesWithBodyWithResponse(gomock.Any(), conversationId, "application/json", gomock.Any()).Return(&clients.PostInternalConversationsConversationIdMessagesResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusConflict},
				Body:         []byte(`{"code":"agent_not_minted","message":"agent identity is not minted yet"}`),
			}, nil)
		}

		t.Run("Should map the 409 to a ConflictError keeping the code", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.AppendMessages(context.Background(), conversationId, request)

			var conflictErr *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictErr)
			assert.Equal(t, "agent_not_minted", conflictErr.Code)
		})
	})

	t.Run("Given the runtime rejects the append", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenAppendingConversationMessagesTestingSuite) {
			suite.mockClient.EXPECT().PostInternalConversationsConversationIdMessagesWithBodyWithResponse(gomock.Any(), conversationId, "application/json", gomock.Any()).Return(&clients.PostInternalConversationsConversationIdMessagesResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError},
				Body:         []byte(`{"code":"boom","message":"runtime down"}`),
			}, nil)
		}

		t.Run("Should return an error", func(t *testing.T) {
			t.Parallel()

			suite := WhenAppendingConversationMessagesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.AppendMessages(context.Background(), conversationId, request)

			assert.Error(t, err)
		})
	})
}
