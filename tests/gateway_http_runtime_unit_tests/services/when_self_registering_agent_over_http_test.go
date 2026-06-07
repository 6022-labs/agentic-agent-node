package services_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services"
	clients_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_http_runtime_mocks/clients_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenSelfRegisteringAgentOverHttpTestingSuite struct {
	sut *services.HttpAgentRegistrationClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenSelfRegisteringAgentOverHttpBeforeEach(t *testing.T) *WhenSelfRegisteringAgentOverHttpTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenSelfRegisteringAgentOverHttpTestingSuite{
		sut:        services.NewHttpAgentRegistrationClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenSelfRegisteringAgentOverHttp(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime accepts the self-registration", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSelfRegisteringAgentOverHttpTestingSuite) {
			suite.mockClient.EXPECT().PostInternalAgentRegistrationSelfWithResponse(gomock.Any()).Return(&clients.PostInternalAgentRegistrationSelfResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusOK},
				Body:         []byte(`{"mintStatus":"submitted","ensStatus":"pending"}`),
			}, nil)
		}

		t.Run("Should decode the registration status payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenSelfRegisteringAgentOverHttpBeforeEach(t)
			initSuite(suite)

			payload, err := suite.sut.Self(context.Background())

			assert.NoError(t, err)
			if assert.NotNil(t, payload) {
				assert.Equal(t, "submitted", payload.MintStatus)
			}
		})
	})

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSelfRegisteringAgentOverHttpTestingSuite) {
			suite.mockClient.EXPECT().PostInternalAgentRegistrationSelfWithResponse(gomock.Any()).Return(nil, errors.New("runtime down"))
		}

		t.Run("Should return an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenSelfRegisteringAgentOverHttpBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Self(context.Background())

			var target *apperrors.UnavailableError
			assert.ErrorAs(t, err, &target)
		})
	})
}
