package services_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/gateway/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services"
	clients_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_http_runtime_mocks/clients_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenCreatingResponseOverHttpTestingSuite struct {
	sut *services.HttpResponseClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenCreatingResponseOverHttpBeforeEach(t *testing.T) *WhenCreatingResponseOverHttpTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenCreatingResponseOverHttpTestingSuite{
		sut:        services.NewHttpResponseClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenCreatingResponseOverHttp(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Truncate(time.Microsecond)
	request := &models.ExternalCreateResponsePayload{
		ConversationId: "conv-1",
		Inputs: []models.ExternalCreateResponseInputPayload{
			{Role: "user", Content: "hello", Timestamp: now},
		},
	}

	t.Run("Given the runtime returns an assistant message", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingResponseOverHttpTestingSuite) {
			suite.mockClient.EXPECT().PostInternalResponsesWithBodyWithResponse(gomock.Any(), "application/json", gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, reader io.Reader, _ ...clients.RequestEditorFn) (*clients.PostInternalResponsesResponse, error) {
					var sent models.ExternalCreateResponsePayload
					raw, _ := io.ReadAll(reader)
					_ = json.Unmarshal(raw, &sent)
					assert.Equal(t, "conv-1", sent.ConversationId)
					return &clients.PostInternalResponsesResponse{
						HTTPResponse: &http.Response{StatusCode: http.StatusOK},
						Body:         []byte(`{"message":{"role":"assistant","content":"hi","timestamp":"2026-01-01T00:00:00Z"}}`),
					}, nil
				})
		}

		t.Run("Should send the conversation and decode the message", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseOverHttpBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Create(context.Background(), request)

			assert.NoError(t, err)
			if assert.NotNil(t, result) && assert.NotNil(t, result.Message) {
				assert.Equal(t, "assistant", result.Message.Role)
				assert.Equal(t, "hi", result.Message.Content)
			}
		})
	})

	t.Run("Given the runtime returns an empty message", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingResponseOverHttpTestingSuite) {
			suite.mockClient.EXPECT().PostInternalResponsesWithBodyWithResponse(gomock.Any(), "application/json", gomock.Any()).
				Return(&clients.PostInternalResponsesResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusOK},
					Body:         []byte(`{}`),
				}, nil)
		}

		t.Run("Should return a result with a nil message", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseOverHttpBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Create(context.Background(), request)

			assert.NoError(t, err)
			if assert.NotNil(t, result) {
				assert.Nil(t, result.Message)
			}
		})
	})

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingResponseOverHttpTestingSuite) {
			suite.mockClient.EXPECT().PostInternalResponsesWithBodyWithResponse(gomock.Any(), "application/json", gomock.Any()).
				Return(nil, errors.New("runtime down"))
		}

		t.Run("Should return the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseOverHttpBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Create(context.Background(), request)

			assert.ErrorContains(t, err, "runtime down")
		})
	})
}
