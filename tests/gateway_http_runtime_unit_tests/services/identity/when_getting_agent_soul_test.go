package services_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	services "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services/identity"
	clients_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_http_runtime_mocks/clients_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenGettingAgentSoulTestingSuite struct {
	sut *services.HttpAgentSoulClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenGettingAgentSoulBeforeEach(t *testing.T) *WhenGettingAgentSoulTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenGettingAgentSoulTestingSuite{
		sut:        services.NewHttpAgentSoulClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenGettingAgentSoul(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentSoulTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalAgentSoulWithResponse(gomock.Any()).
				Return(nil, errors.New("connection refused"))
		}

		t.Run("Should map the transport error onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentSoulBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Get(context.Background())

			var target *apperrors.UnavailableError
			if assert.ErrorAs(t, err, &target) {
				assert.Equal(t, "runtime_unavailable", target.Code)
			}
		})
	})

	t.Run("Given the runtime reports no identity", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentSoulTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalAgentSoulWithResponse(gomock.Any()).
				Return(&clients.GetInternalAgentSoulResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusNotFound},
					Body:         []byte(`{"code":"identity_not_found","message":"agent identity not found"}`),
				}, nil)
		}

		t.Run("Should surface the not-found sentinel", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentSoulBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Get(context.Background())

			assert.ErrorIs(t, err, identity_interfaces.ErrAgentSoulNotFound)
		})
	})

	t.Run("Given the runtime answers 500", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentSoulTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalAgentSoulWithResponse(gomock.Any()).
				Return(&clients.GetInternalAgentSoulResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError},
					Body:         []byte(`{"code":"x","field":"f","message":"m"}`),
				}, nil)
		}

		t.Run("Should map the 500 onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentSoulBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Get(context.Background())

			var target *apperrors.UnavailableError
			if assert.ErrorAs(t, err, &target) {
				assert.Equal(t, "runtime_unavailable", target.Code)
			}
		})
	})

	t.Run("Given the runtime returns the soul", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentSoulTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalAgentSoulWithResponse(gomock.Any()).
				Return(&clients.GetInternalAgentSoulResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusOK},
					Body:         []byte(`{"soul":"persona text"}`),
				}, nil)
		}

		t.Run("Should decode the soul payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentSoulBeforeEach(t)
			initSuite(suite)

			payload, err := suite.sut.Get(context.Background())

			assert.NoError(t, err)
			if assert.NotNil(t, payload) {
				assert.Equal(t, "persona text", payload.Soul)
			}
		})
	})
}
