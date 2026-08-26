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

type WhenCreatingConversationTestingSuite struct {
	sut *services.HttpConversationClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenCreatingConversationBeforeEach(t *testing.T) *WhenCreatingConversationTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenCreatingConversationTestingSuite{
		sut:        services.NewHttpConversationClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenCreatingConversation(t *testing.T) {
	t.Parallel()

	request := &models.ExternalCreateConversationPayload{Name: "ERC20 mint audit"}

	t.Run("Given the runtime creates the session", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingConversationTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalConversationsWithBodyWithResponse(gomock.Any(), "application/json", gomock.Any()).
				DoAndReturn(
					func(_ context.Context, _ string, reader io.Reader, _ ...clients.RequestEditorFn) (*clients.PostInternalConversationsResponse, error) {
						raw, err := io.ReadAll(reader)
						assert.NoError(t, err)

						var sent models.ExternalCreateConversationPayload
						assert.NoError(t, json.Unmarshal(raw, &sent))
						assert.Equal(t, "ERC20 mint audit", sent.Name)

						return &clients.PostInternalConversationsResponse{
							HTTPResponse: &http.Response{StatusCode: http.StatusCreated},
							Body: []byte(
								`{"conversationId":"c1","name":"ERC20 mint audit","messageCount":0,` +
									`"createdAt":"2026-01-01T00:00:00Z","updatedAt":"2026-01-01T00:00:00Z"}`,
							),
						}, nil
					})
		}

		t.Run("Should decode the created conversation", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingConversationBeforeEach(t)
			initSuite(suite)

			payload, err := suite.sut.CreateConversation(context.Background(), request)

			assert.NoError(t, err)
			if assert.NotNil(t, payload) {
				assert.Equal(t, "c1", payload.ConversationId)
				assert.Equal(t, "ERC20 mint audit", lo.FromPtr(payload.Name))
			}
		})
	})

	t.Run("Given the runtime reports the name taken", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingConversationTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalConversationsWithBodyWithResponse(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&clients.PostInternalConversationsResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusConflict},
					Body:         []byte(`{"code":"conversation_name_taken","message":"already used"}`),
				}, nil)
		}

		t.Run("Should map the 409 onto a ConflictError", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingConversationBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.CreateConversation(context.Background(), request)

			var conflictErr *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictErr)
			assert.Equal(t, "conversation_name_taken", conflictErr.Code)
		})
	})
}
