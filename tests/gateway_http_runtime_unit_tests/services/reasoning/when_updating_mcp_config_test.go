package services_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"

	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/models"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	services "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services/reasoning"
	clients_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_http_runtime_mocks/clients_mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenUpdatingMCPConfigTestingSuite struct {
	sut *services.HttpMCPConfigClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenUpdatingMCPConfigBeforeEach(t *testing.T) *WhenUpdatingMCPConfigTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenUpdatingMCPConfigTestingSuite{
		sut:        services.NewHttpMCPConfigClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenUpdatingMCPConfig(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()

		initSuite := func(suite *WhenUpdatingMCPConfigTestingSuite) {
			suite.mockClient.EXPECT().
				PutInternalConfigMcpIdWithBodyWithResponse(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, errors.New("runtime down"))
		}

		t.Run("Should map the transport error onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingMCPConfigBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Update(context.Background(), id, &models.ExternalUpdateMCPConnectionPayload{
				Name: "renamed",
				Url:  "https://mcp.example",
			})

			var target *apperrors.UnavailableError
			if assert.ErrorAs(t, err, &target) {
				assert.Equal(t, "runtime_unavailable", target.Code)
			}
		})
	})

	t.Run("Given the runtime answers 400", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()

		initSuite := func(suite *WhenUpdatingMCPConfigTestingSuite) {
			suite.mockClient.EXPECT().
				PutInternalConfigMcpIdWithBodyWithResponse(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&clients.PutInternalConfigMcpIdResponse{HTTPResponse: &http.Response{StatusCode: http.StatusBadRequest}, Body: []byte(`{"code":"x","field":"f","message":"m"}`)}, nil)
		}

		t.Run("Should map the 400 onto a ValidationError", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingMCPConfigBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Update(context.Background(), id, &models.ExternalUpdateMCPConnectionPayload{
				Name: "renamed",
				Url:  "https://mcp.example",
			})

			var target *apperrors.ValidationError
			assert.ErrorAs(t, err, &target)
		})
	})

	t.Run("Given the connection does not exist", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()

		initSuite := func(suite *WhenUpdatingMCPConfigTestingSuite) {
			suite.mockClient.EXPECT().
				PutInternalConfigMcpIdWithBodyWithResponse(gomock.Any(), id.String(), "application/json", gomock.Any()).
				Return(&clients.PutInternalConfigMcpIdResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusNotFound},
					Body:         []byte(`{"code":"not_found"}`),
				}, nil)
		}

		t.Run("Should return ErrMCPConnectionNotFound", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingMCPConfigBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Update(context.Background(), id, &models.ExternalUpdateMCPConnectionPayload{
				Name: "renamed",
				Url:  "https://mcp.example",
			})

			assert.ErrorIs(t, err, gateway_interfaces.ErrMCPConnectionNotFound)
		})
	})

	t.Run("Given the runtime answers 409", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()

		initSuite := func(suite *WhenUpdatingMCPConfigTestingSuite) {
			suite.mockClient.EXPECT().
				PutInternalConfigMcpIdWithBodyWithResponse(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&clients.PutInternalConfigMcpIdResponse{HTTPResponse: &http.Response{StatusCode: http.StatusConflict}, Body: []byte(`{"code":"x","field":"f","message":"m"}`)}, nil)
		}

		t.Run("Should map the 409 onto a ConflictError", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingMCPConfigBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Update(context.Background(), id, &models.ExternalUpdateMCPConnectionPayload{
				Name: "renamed",
				Url:  "https://mcp.example",
			})

			var target *apperrors.ConflictError
			assert.ErrorAs(t, err, &target)
		})
	})

	t.Run("Given the runtime answers 500", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()

		initSuite := func(suite *WhenUpdatingMCPConfigTestingSuite) {
			suite.mockClient.EXPECT().
				PutInternalConfigMcpIdWithBodyWithResponse(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&clients.PutInternalConfigMcpIdResponse{HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError}, Body: []byte(`{"code":"x","field":"f","message":"m"}`)}, nil)
		}

		t.Run("Should map the 500 onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingMCPConfigBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Update(context.Background(), id, &models.ExternalUpdateMCPConnectionPayload{
				Name: "renamed",
				Url:  "https://mcp.example",
			})

			var target *apperrors.UnavailableError
			if assert.ErrorAs(t, err, &target) {
				assert.Equal(t, "runtime_unavailable", target.Code)
			}
		})
	})

	t.Run("Given the runtime accepts the update", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()

		initSuite := func(suite *WhenUpdatingMCPConfigTestingSuite) {
			suite.mockClient.EXPECT().
				PutInternalConfigMcpIdWithBodyWithResponse(gomock.Any(), id.String(), "application/json", gomock.Any()).
				DoAndReturn(
					func(_ context.Context, _ string, _ string, reader io.Reader, _ ...clients.RequestEditorFn) (*clients.PutInternalConfigMcpIdResponse, error) {
						raw, err := io.ReadAll(reader)
						assert.NoError(t, err)

						var sent models.ExternalUpdateMCPConnectionPayload
						assert.NoError(t, json.Unmarshal(raw, &sent))
						assert.Equal(t, "renamed", sent.Name)

						return &clients.PutInternalConfigMcpIdResponse{
							HTTPResponse: &http.Response{StatusCode: http.StatusOK},
							Body: []byte(
								`{"id":"mcp-1","name":"renamed","url":"https://mcp.example","headerAuthKeys":[]}`,
							),
						}, nil
					})
		}

		t.Run("Should decode the updated MCP connection payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingMCPConfigBeforeEach(t)
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
}
