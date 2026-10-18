package services_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	services "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services/identity"
	clients_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_http_runtime_mocks/clients_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenCheckingAgentEnsTestingSuite struct {
	sut *services.HttpAgentRegistrationClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenCheckingAgentEnsBeforeEach(t *testing.T) *WhenCheckingAgentEnsTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenCheckingAgentEnsTestingSuite{
		sut:        services.NewHttpAgentRegistrationClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenCheckingAgentEns(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCheckingAgentEnsTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalAgentRegistrationEnsCheckWithResponse(gomock.Any()).
				Return(nil, errors.New("runtime down"))
		}

		t.Run("Should map the transport error onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingAgentEnsBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.CheckEns(context.Background())

			var target *apperrors.UnavailableError
			if assert.ErrorAs(t, err, &target) {
				assert.Equal(t, "runtime_unavailable", target.Code)
			}
		})
	})

	t.Run("Given the runtime answers 409", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCheckingAgentEnsTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalAgentRegistrationEnsCheckWithResponse(gomock.Any()).
				Return(&clients.PostInternalAgentRegistrationEnsCheckResponse{HTTPResponse: &http.Response{StatusCode: http.StatusConflict}, Body: []byte(`{"code":"x","field":"f","message":"m"}`)}, nil)
		}

		t.Run("Should map the 409 onto a ConflictError", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingAgentEnsBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.CheckEns(context.Background())

			var target *apperrors.ConflictError
			assert.ErrorAs(t, err, &target)
		})
	})

	t.Run("Given the runtime answers 500", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCheckingAgentEnsTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalAgentRegistrationEnsCheckWithResponse(gomock.Any()).
				Return(&clients.PostInternalAgentRegistrationEnsCheckResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError},
					Body:         []byte(`{"code":"boom"}`),
				}, nil)
		}

		t.Run("Should return an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingAgentEnsBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.CheckEns(context.Background())

			var target *apperrors.UnavailableError
			assert.ErrorAs(t, err, &target)
		})
	})

	t.Run("Given the runtime returns the ENS status", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCheckingAgentEnsTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalAgentRegistrationEnsCheckWithResponse(gomock.Any()).
				Return(&clients.PostInternalAgentRegistrationEnsCheckResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusOK},
					Body:         []byte(`{"mintStatus":"minted","ensStatus":"published"}`),
				}, nil)
		}

		t.Run("Should decode the registration status payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingAgentEnsBeforeEach(t)
			initSuite(suite)

			payload, err := suite.sut.CheckEns(context.Background())

			assert.NoError(t, err)
			if assert.NotNil(t, payload) {
				assert.Equal(t, "published", payload.EnsStatus)
			}
		})
	})
}
