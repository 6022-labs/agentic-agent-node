package services_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	services "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services/payment"
	clients_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_http_runtime_mocks/clients_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenGettingPaymentFacilitatorsTestingSuite struct {
	sut *services.HttpPaymentFacilitatorClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenGettingPaymentFacilitatorsBeforeEach(t *testing.T) *WhenGettingPaymentFacilitatorsTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenGettingPaymentFacilitatorsTestingSuite{
		sut:        services.NewHttpPaymentFacilitatorClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenGettingPaymentFacilitators(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingPaymentFacilitatorsTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalConfigPaymentFacilitatorsWithResponse(gomock.Any()).
				Return(nil, errors.New("runtime down"))
		}

		t.Run("Should return an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingPaymentFacilitatorsBeforeEach(t)
			initSuite(suite)

			facilitators, err := suite.sut.Get(context.Background())

			assert.Nil(t, facilitators)
			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Equal(t, "runtime_unavailable", unavailableErr.Code)
		})
	})

	t.Run("Given the runtime answers 500", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingPaymentFacilitatorsTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalConfigPaymentFacilitatorsWithResponse(gomock.Any()).
				Return(&clients.GetInternalConfigPaymentFacilitatorsResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError},
					Body:         []byte(`{"code":"boom"}`),
				}, nil)
		}

		t.Run("Should return an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingPaymentFacilitatorsBeforeEach(t)
			initSuite(suite)

			facilitators, err := suite.sut.Get(context.Background())

			assert.Nil(t, facilitators)
			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Equal(t, "runtime_unavailable", unavailableErr.Code)
		})
	})

	t.Run("Given the runtime returns the facilitators", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingPaymentFacilitatorsTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalConfigPaymentFacilitatorsWithResponse(gomock.Any()).
				Return(&clients.GetInternalConfigPaymentFacilitatorsResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusOK},
					Body: []byte(
						`{"facilitators":[{"network":"eip155:80002",` +
							`"baseUrl":"https://x402-amoy.polygon.technology","maxTimeoutSeconds":90}]}`,
					),
				}, nil)
		}

		t.Run("Should decode them into a network → facilitator map", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingPaymentFacilitatorsBeforeEach(t)
			initSuite(suite)

			facilitators, err := suite.sut.Get(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, "https://x402-amoy.polygon.technology", facilitators["eip155:80002"].BaseUrl)
			assert.Equal(t, 90, facilitators["eip155:80002"].MaxTimeoutSeconds)
		})
	})
}
