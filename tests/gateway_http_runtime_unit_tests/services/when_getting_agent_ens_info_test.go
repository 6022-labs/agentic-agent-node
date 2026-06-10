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

type WhenGettingAgentEnsInfoTestingSuite struct {
	sut *services.HttpAgentRegistrationClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenGettingAgentEnsInfoBeforeEach(t *testing.T) *WhenGettingAgentEnsInfoTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenGettingAgentEnsInfoTestingSuite{
		sut:        services.NewHttpAgentRegistrationClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenGettingAgentEnsInfo(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime returns the ENS calldata", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentEnsInfoTestingSuite) {
			suite.mockClient.EXPECT().GetInternalAgentRegistrationEnsInfoWithResponse(gomock.Any()).Return(&clients.GetInternalAgentRegistrationEnsInfoResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusOK},
				Body:         []byte(`{"collectionAddress":"0xabc","tokenId":7,"records":[]}`),
			}, nil)
		}

		t.Run("Should decode the ENS info payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentEnsInfoBeforeEach(t)
			initSuite(suite)

			payload, err := suite.sut.EnsInfo(context.Background())

			assert.NoError(t, err)
			if assert.NotNil(t, payload) {
				assert.Equal(t, "0xabc", payload.CollectionAddress)
				assert.Equal(t, uint64(7), payload.TokenId)
			}
		})
	})

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentEnsInfoTestingSuite) {
			suite.mockClient.EXPECT().GetInternalAgentRegistrationEnsInfoWithResponse(gomock.Any()).Return(nil, errors.New("runtime down"))
		}

		t.Run("Should return an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentEnsInfoBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.EnsInfo(context.Background())

			var target *apperrors.UnavailableError
			assert.ErrorAs(t, err, &target)
		})
	})
}
