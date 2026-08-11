package services_test

import (
	"context"
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

	t.Run("Given the runtime fails to check the ENS status", func(t *testing.T) {
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
}
