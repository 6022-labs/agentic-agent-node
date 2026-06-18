package services_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	services "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services/reasoning"
	clients_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_http_runtime_mocks/clients_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenCreatingResponseTestingSuite struct {
	sut *services.HttpResponseClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenCreatingResponseBeforeEach(t *testing.T) *WhenCreatingResponseTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenCreatingResponseTestingSuite{
		sut:        services.NewHttpResponseClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenCreatingResponse(t *testing.T) {
	t.Parallel()

	request := &models.ExternalCreateResponsePayload{
		PreviousResponseId: "conv-1",
		Input: []models.ExternalInputMessagePayload{
			{Type: "message", Role: "user", Content: "hello"},
		},
	}

	t.Run("Given the runtime returns a completed response", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingResponseTestingSuite) {
			suite.mockClient.EXPECT().PostInternalResponsesWithBodyWithResponse(gomock.Any(), "application/json", gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, reader io.Reader, _ ...clients.RequestEditorFn) (*clients.PostInternalResponsesResponse, error) {
					var sent models.ExternalCreateResponsePayload
					raw, _ := io.ReadAll(reader)
					_ = json.Unmarshal(raw, &sent)
					assert.Equal(t, "conv-1", sent.PreviousResponseId)
					return &clients.PostInternalResponsesResponse{
						HTTPResponse: &http.Response{StatusCode: http.StatusOK},
						Body:         []byte(`{"id":"conv-1","object":"response","status":"completed","output":[{"type":"message","role":"assistant","content":[{"type":"output_text","text":"hi"}]}]}`),
					}, nil
				})
		}

		t.Run("Should send the request and decode the OpenAI response", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Create(context.Background(), request)

			assert.NoError(t, err)
			if assert.NotNil(t, result) {
				assert.Equal(t, "conv-1", result.Id)
				assert.Equal(t, "completed", result.Status)
				if assert.Len(t, result.Output, 1) && assert.Len(t, result.Output[0].Content, 1) {
					assert.Equal(t, "hi", result.Output[0].Content[0].Text)
				}
			}
		})
	})

	t.Run("Given the runtime returns an incomplete response", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingResponseTestingSuite) {
			suite.mockClient.EXPECT().PostInternalResponsesWithBodyWithResponse(gomock.Any(), "application/json", gomock.Any()).
				Return(&clients.PostInternalResponsesResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusOK},
					Body:         []byte(`{"id":"conv-1","object":"response","status":"incomplete","output":[]}`),
				}, nil)
		}

		t.Run("Should return a result with no output", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Create(context.Background(), request)

			assert.NoError(t, err)
			if assert.NotNil(t, result) {
				assert.Equal(t, "incomplete", result.Status)
				assert.Empty(t, result.Output)
			}
		})
	})

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingResponseTestingSuite) {
			suite.mockClient.EXPECT().PostInternalResponsesWithBodyWithResponse(gomock.Any(), "application/json", gomock.Any()).
				Return(nil, errors.New("runtime down"))
		}

		t.Run("Should return the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Create(context.Background(), request)

			assert.ErrorContains(t, err, "runtime down")
		})
	})
}
