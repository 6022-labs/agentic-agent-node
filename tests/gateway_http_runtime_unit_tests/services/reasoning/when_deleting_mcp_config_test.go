package services_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"

	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	services "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services/reasoning"
	clients_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_http_runtime_mocks/clients_mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenDeletingMCPConfigTestingSuite struct {
	sut *services.HttpMCPConfigClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenDeletingMCPConfigBeforeEach(t *testing.T) *WhenDeletingMCPConfigTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenDeletingMCPConfigTestingSuite{
		sut:        services.NewHttpMCPConfigClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenDeletingMCPConfig(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()

		initSuite := func(suite *WhenDeletingMCPConfigTestingSuite) {
			suite.mockClient.EXPECT().
				DeleteInternalConfigMcpIdWithResponse(gomock.Any(), gomock.Any()).
				Return(nil, errors.New("runtime down"))
		}

		t.Run("Should map the transport error onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingMCPConfigBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Delete(context.Background(), id)

			var target *apperrors.UnavailableError
			if assert.ErrorAs(t, err, &target) {
				assert.Equal(t, "runtime_unavailable", target.Code)
			}
		})
	})

	t.Run("Given the runtime answers 400", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()

		initSuite := func(suite *WhenDeletingMCPConfigTestingSuite) {
			suite.mockClient.EXPECT().
				DeleteInternalConfigMcpIdWithResponse(gomock.Any(), gomock.Any()).
				Return(&clients.DeleteInternalConfigMcpIdResponse{HTTPResponse: &http.Response{StatusCode: http.StatusBadRequest}, Body: []byte(`{"code":"x","field":"f","message":"m"}`)}, nil)
		}

		t.Run("Should map the 400 onto a ValidationError", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingMCPConfigBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Delete(context.Background(), id)

			var target *apperrors.ValidationError
			assert.ErrorAs(t, err, &target)
		})
	})

	t.Run("Given the runtime answers 500", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()

		initSuite := func(suite *WhenDeletingMCPConfigTestingSuite) {
			suite.mockClient.EXPECT().
				DeleteInternalConfigMcpIdWithResponse(gomock.Any(), id.String()).
				Return(&clients.DeleteInternalConfigMcpIdResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError},
					Body:         []byte(`{"code":"boom"}`),
				}, nil)
		}

		t.Run("Should map the 500 onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingMCPConfigBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Delete(context.Background(), id)

			var target *apperrors.UnavailableError
			if assert.ErrorAs(t, err, &target) {
				assert.Equal(t, "runtime_unavailable", target.Code)
			}
		})
	})

	t.Run("Given the runtime deletes the connection", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()

		initSuite := func(suite *WhenDeletingMCPConfigTestingSuite) {
			suite.mockClient.EXPECT().
				DeleteInternalConfigMcpIdWithResponse(gomock.Any(), id.String()).
				Return(&clients.DeleteInternalConfigMcpIdResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusNoContent},
				}, nil)
		}

		t.Run("Should not return an error", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingMCPConfigBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Delete(context.Background(), id)

			assert.NoError(t, err)
		})
	})
}
