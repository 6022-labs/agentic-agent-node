package services_test

import (
	"context"
	"encoding/json"
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

type WhenCreatingMCPConfigTestingSuite struct {
	sut *services.HttpMCPConfigClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenCreatingMCPConfigBeforeEach(t *testing.T) *WhenCreatingMCPConfigTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenCreatingMCPConfigTestingSuite{
		sut:        services.NewHttpMCPConfigClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenCreatingMCPConfig(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime accepts the new connection", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingMCPConfigTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalConfigMcpWithBodyWithResponse(gomock.Any(), "application/json", gomock.Any()).
				DoAndReturn(
					func(_ context.Context, _ string, reader io.Reader, _ ...clients.RequestEditorFn) (*clients.PostInternalConfigMcpResponse, error) {
						raw, err := io.ReadAll(reader)
						assert.NoError(t, err)

						var sent models.ExternalCreateMCPConnectionPayload
						assert.NoError(t, json.Unmarshal(raw, &sent))
						assert.Equal(t, "tools", sent.Name)
						assert.Equal(t, "Bearer x", sent.HeaderAuth["Authorization"])

						return &clients.PostInternalConfigMcpResponse{
							HTTPResponse: &http.Response{StatusCode: http.StatusOK},
							Body: []byte(
								`{"id":"mcp-1","name":"tools","url":"https://mcp.example","headerAuthKeys":["Authorization"]}`,
							),
						}, nil
					})
		}

		t.Run("Should decode the created MCP connection payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingMCPConfigBeforeEach(t)
			initSuite(suite)

			payload, err := suite.sut.Create(context.Background(), &models.ExternalCreateMCPConnectionPayload{
				Name:       "tools",
				Url:        "https://mcp.example",
				HeaderAuth: map[string]string{"Authorization": "Bearer x"},
			})

			assert.NoError(t, err)
			if assert.NotNil(t, payload) {
				assert.Equal(t, "mcp-1", payload.Id)
				assert.Equal(t, []string{"Authorization"}, payload.HeaderAuthKeys)
			}
		})
	})

	t.Run("Given the runtime rejects the new connection", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingMCPConfigTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalConfigMcpWithBodyWithResponse(gomock.Any(), "application/json", gomock.Any()).
				Return(&clients.PostInternalConfigMcpResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError},
					Body:         []byte(`{"code":"boom"}`),
				}, nil)
		}

		t.Run("Should return an error", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingMCPConfigBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Create(context.Background(), &models.ExternalCreateMCPConnectionPayload{
				Name:       "tools",
				Url:        "https://mcp.example",
				HeaderAuth: map[string]string{"Authorization": "Bearer x"},
			})

			assert.Error(t, err)
		})
	})
}
