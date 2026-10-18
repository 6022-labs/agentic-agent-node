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

type WhenCheckingAgentMintTestingSuite struct {
	sut *services.HttpAgentRegistrationClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenCheckingAgentMintBeforeEach(t *testing.T) *WhenCheckingAgentMintTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenCheckingAgentMintTestingSuite{
		sut:        services.NewHttpAgentRegistrationClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenCheckingAgentMint(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCheckingAgentMintTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalAgentRegistrationMintCheckWithResponse(gomock.Any()).
				Return(nil, errors.New("runtime down"))
		}

		t.Run("Should map the transport error onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingAgentMintBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.CheckMint(context.Background())

			var target *apperrors.UnavailableError
			if assert.ErrorAs(t, err, &target) {
				assert.Equal(t, "runtime_unavailable", target.Code)
			}
		})
	})

	t.Run("Given the runtime answers 404", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCheckingAgentMintTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalAgentRegistrationMintCheckWithResponse(gomock.Any()).
				Return(&clients.PostInternalAgentRegistrationMintCheckResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusNotFound},
					Body:         []byte(`{"code":"x","field":"f","message":"m"}`),
				}, nil)
		}

		t.Run("Should map the 404 onto a NotFoundError", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingAgentMintBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.CheckMint(context.Background())

			var target *apperrors.NotFoundError
			assert.ErrorAs(t, err, &target)
		})
	})

	t.Run("Given a mint is already in flight", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCheckingAgentMintTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalAgentRegistrationMintCheckWithResponse(gomock.Any()).
				Return(&clients.PostInternalAgentRegistrationMintCheckResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusConflict},
					Body:         []byte(`{"code":"conflict","message":"already minting"}`),
				}, nil)
		}

		t.Run("Should map the 409 onto a ConflictError", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingAgentMintBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.CheckMint(context.Background())

			var target *apperrors.ConflictError
			assert.ErrorAs(t, err, &target)
		})
	})

	t.Run("Given the runtime answers 500", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCheckingAgentMintTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalAgentRegistrationMintCheckWithResponse(gomock.Any()).
				Return(&clients.PostInternalAgentRegistrationMintCheckResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError},
					Body:         []byte(`{"code":"x","field":"f","message":"m"}`),
				}, nil)
		}

		t.Run("Should map the 500 onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingAgentMintBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.CheckMint(context.Background())

			var target *apperrors.UnavailableError
			if assert.ErrorAs(t, err, &target) {
				assert.Equal(t, "runtime_unavailable", target.Code)
			}
		})
	})

	t.Run("Given the runtime returns the mint status", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCheckingAgentMintTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalAgentRegistrationMintCheckWithResponse(gomock.Any()).
				Return(&clients.PostInternalAgentRegistrationMintCheckResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusOK},
					Body:         []byte(`{"mintStatus":"minted","ensStatus":"pending"}`),
				}, nil)
		}

		t.Run("Should decode the registration status payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingAgentMintBeforeEach(t)
			initSuite(suite)

			payload, err := suite.sut.CheckMint(context.Background())

			assert.NoError(t, err)
			if assert.NotNil(t, payload) {
				assert.Equal(t, "minted", payload.MintStatus)
			}
		})
	})
}
