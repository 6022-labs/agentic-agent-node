package services_test

import (
	"context"
	"encoding/json"
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

	t.Run("Given the runtime does not know the conversation", func(t *testing.T) {
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
}
