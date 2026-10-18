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

type WhenGettingAgentMintInfoTestingSuite struct {
	sut *services.HttpAgentRegistrationClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenGettingAgentMintInfoBeforeEach(t *testing.T) *WhenGettingAgentMintInfoTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenGettingAgentMintInfoTestingSuite{
		sut:        services.NewHttpAgentRegistrationClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenGettingAgentMintInfo(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentMintInfoTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalAgentRegistrationMintInfoWithResponse(gomock.Any()).
				Return(nil, errors.New("runtime down"))
		}

		t.Run("Should map the transport error onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentMintInfoBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.MintInfo(context.Background())

			var target *apperrors.UnavailableError
			if assert.ErrorAs(t, err, &target) {
				assert.Equal(t, "runtime_unavailable", target.Code)
			}
		})
	})

	t.Run("Given the runtime answers 404", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentMintInfoTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalAgentRegistrationMintInfoWithResponse(gomock.Any()).
				Return(&clients.GetInternalAgentRegistrationMintInfoResponse{HTTPResponse: &http.Response{StatusCode: http.StatusNotFound}, Body: []byte(`{"code":"x","field":"f","message":"m"}`)}, nil)
		}

		t.Run("Should map the 404 onto a NotFoundError", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentMintInfoBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.MintInfo(context.Background())

			var target *apperrors.NotFoundError
			assert.ErrorAs(t, err, &target)
		})
	})

	t.Run("Given the runtime answers 409", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentMintInfoTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalAgentRegistrationMintInfoWithResponse(gomock.Any()).
				Return(&clients.GetInternalAgentRegistrationMintInfoResponse{HTTPResponse: &http.Response{StatusCode: http.StatusConflict}, Body: []byte(`{"code":"x","field":"f","message":"m"}`)}, nil)
		}

		t.Run("Should map the 409 onto a ConflictError", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentMintInfoBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.MintInfo(context.Background())

			var target *apperrors.ConflictError
			assert.ErrorAs(t, err, &target)
		})
	})

	t.Run("Given the runtime answers 500", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentMintInfoTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalAgentRegistrationMintInfoWithResponse(gomock.Any()).
				Return(&clients.GetInternalAgentRegistrationMintInfoResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError},
					Body:         []byte(`{"code":"boom"}`),
				}, nil)
		}

		t.Run("Should return an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentMintInfoBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.MintInfo(context.Background())

			var target *apperrors.UnavailableError
			assert.ErrorAs(t, err, &target)
		})
	})

	t.Run("Given the runtime returns the mint calldata", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentMintInfoTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalAgentRegistrationMintInfoWithResponse(gomock.Any()).
				Return(&clients.GetInternalAgentRegistrationMintInfoResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusOK},
					Body: []byte(
						`{"collectionAddress":"0xabc","to":"0xowner","role":"agent","name":"bob","wallets":[]}`,
					),
				}, nil)
		}

		t.Run("Should decode the mint info payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentMintInfoBeforeEach(t)
			initSuite(suite)

			payload, err := suite.sut.MintInfo(context.Background())

			assert.NoError(t, err)
			if assert.NotNil(t, payload) {
				assert.Equal(t, "0xowner", payload.To)
				assert.Equal(t, "bob", payload.Name)
			}
		})
	})
}
