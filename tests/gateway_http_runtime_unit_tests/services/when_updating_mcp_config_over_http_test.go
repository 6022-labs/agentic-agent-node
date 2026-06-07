package services_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway/models"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services"
	clients_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_http_runtime_mocks/clients_mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenUpdatingMCPConfigOverHttpTestingSuite struct {
	sut *services.HttpMCPConfigClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenUpdatingMCPConfigOverHttpBeforeEach(t *testing.T) *WhenUpdatingMCPConfigOverHttpTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenUpdatingMCPConfigOverHttpTestingSuite{
		sut:        services.NewHttpMCPConfigClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenUpdatingMCPConfigOverHttp(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime accepts the update", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()

		initSuite := func(suite *WhenUpdatingMCPConfigOverHttpTestingSuite) {
			suite.mockClient.EXPECT().PutInternalConfigMcpIdWithBodyWithResponse(gomock.Any(), id.String(), "application/json", gomock.Any()).DoAndReturn(
				func(_ context.Context, _ string, _ string, reader io.Reader, _ ...clients.RequestEditorFn) (*clients.PutInternalConfigMcpIdResponse, error) {
					raw, err := io.ReadAll(reader)
					assert.NoError(t, err)

					var sent models.ExternalUpdateMCPConnectionPayload
					assert.NoError(t, json.Unmarshal(raw, &sent))
					assert.Equal(t, "renamed", sent.Name)

					return &clients.PutInternalConfigMcpIdResponse{
						HTTPResponse: &http.Response{StatusCode: http.StatusOK},
						Body:         []byte(`{"id":"mcp-1","name":"renamed","url":"https://mcp.example","headerAuthKeys":[]}`),
					}, nil
				})
		}

		t.Run("Should decode the updated MCP connection payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingMCPConfigOverHttpBeforeEach(t)
			initSuite(suite)

			payload, err := suite.sut.Update(context.Background(), id, &models.ExternalUpdateMCPConnectionPayload{
				Name: "renamed",
				Url:  "https://mcp.example",
			})

			assert.NoError(t, err)
			if assert.NotNil(t, payload) {
				assert.Equal(t, "renamed", payload.Name)
			}
		})
	})

	t.Run("Given updating a missing connection", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()

		initSuite := func(suite *WhenUpdatingMCPConfigOverHttpTestingSuite) {
			suite.mockClient.EXPECT().PutInternalConfigMcpIdWithBodyWithResponse(gomock.Any(), id.String(), "application/json", gomock.Any()).Return(&clients.PutInternalConfigMcpIdResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusNotFound},
				Body:         []byte(`{"code":"not_found"}`),
			}, nil)
		}

		t.Run("Should return ErrMCPConnectionNotFound", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingMCPConfigOverHttpBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Update(context.Background(), id, &models.ExternalUpdateMCPConnectionPayload{
				Name: "renamed",
				Url:  "https://mcp.example",
			})

			assert.ErrorIs(t, err, gateway_interfaces.ErrMCPConnectionNotFound)
		})
	})
}
