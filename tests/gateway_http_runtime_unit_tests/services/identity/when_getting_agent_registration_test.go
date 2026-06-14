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

type WhenGettingAgentRegistrationTestingSuite struct {
	sut *services.HttpAgentRegistrationClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenGettingAgentRegistrationBeforeEach(t *testing.T) *WhenGettingAgentRegistrationTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenGettingAgentRegistrationTestingSuite{
		sut:        services.NewHttpAgentRegistrationClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenGettingAgentRegistration(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime returns the registration draft", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentRegistrationTestingSuite) {
			suite.mockClient.EXPECT().GetInternalAgentRegistrationWithResponse(gomock.Any()).Return(&clients.GetInternalAgentRegistrationResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusOK},
				Body:         []byte(`{"collectionAddress":"0xabc","name":"bob","role":"agent","owner":"0xowner"}`),
			}, nil)
		}

		t.Run("Should decode the registration payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentRegistrationBeforeEach(t)
			initSuite(suite)

			payload, err := suite.sut.Get(context.Background())

			assert.NoError(t, err)
			if assert.NotNil(t, payload) {
				assert.Equal(t, "bob", payload.Name)
			}
		})
	})

	t.Run("Given no registration draft exists", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentRegistrationTestingSuite) {
			suite.mockClient.EXPECT().GetInternalAgentRegistrationWithResponse(gomock.Any()).Return(&clients.GetInternalAgentRegistrationResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusNotFound},
				Body:         []byte(`{"code":"not_found","message":"no draft"}`),
			}, nil)
		}

		t.Run("Should return a NotFoundError", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentRegistrationBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Get(context.Background())

			var target *apperrors.NotFoundError
			assert.ErrorAs(t, err, &target)
		})
	})
}
