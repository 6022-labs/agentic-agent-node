package services_test

import (
	"context"
	"net/http"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services"
	clients_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_http_runtime_mocks/clients_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenCheckingAgentMintOverHttpTestingSuite struct {
	sut *services.HttpAgentRegistrationClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenCheckingAgentMintOverHttpBeforeEach(t *testing.T) *WhenCheckingAgentMintOverHttpTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenCheckingAgentMintOverHttpTestingSuite{
		sut:        services.NewHttpAgentRegistrationClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenCheckingAgentMintOverHttp(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime returns the mint status", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCheckingAgentMintOverHttpTestingSuite) {
			suite.mockClient.EXPECT().PostInternalAgentRegistrationMintCheckWithResponse(gomock.Any()).Return(&clients.PostInternalAgentRegistrationMintCheckResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusOK},
				Body:         []byte(`{"mintStatus":"minted","ensStatus":"pending"}`),
			}, nil)
		}

		t.Run("Should decode the registration status payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingAgentMintOverHttpBeforeEach(t)
			initSuite(suite)

			payload, err := suite.sut.CheckMint(context.Background())

			assert.NoError(t, err)
			if assert.NotNil(t, payload) {
				assert.Equal(t, "minted", payload.MintStatus)
			}
		})
	})

	t.Run("Given a mint is already in flight", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCheckingAgentMintOverHttpTestingSuite) {
			suite.mockClient.EXPECT().PostInternalAgentRegistrationMintCheckWithResponse(gomock.Any()).Return(&clients.PostInternalAgentRegistrationMintCheckResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusConflict},
				Body:         []byte(`{"code":"conflict","message":"already minting"}`),
			}, nil)
		}

		t.Run("Should return a ConflictError", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingAgentMintOverHttpBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.CheckMint(context.Background())

			var target *apperrors.ConflictError
			assert.ErrorAs(t, err, &target)
		})
	})
}
