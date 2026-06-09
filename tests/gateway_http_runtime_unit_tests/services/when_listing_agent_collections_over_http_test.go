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

type WhenListingAgentCollectionsOverHttpTestingSuite struct {
	sut *services.HttpAgentCollectionsClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenListingAgentCollectionsOverHttpBeforeEach(t *testing.T) *WhenListingAgentCollectionsOverHttpTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenListingAgentCollectionsOverHttpTestingSuite{
		sut:        services.NewHttpAgentCollectionsClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenListingAgentCollectionsOverHttp(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime returns the collections catalog", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingAgentCollectionsOverHttpTestingSuite) {
			suite.mockClient.EXPECT().GetInternalAgentCollectionsWithResponse(gomock.Any()).Return(&clients.GetInternalAgentCollectionsResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusOK},
				Body:         []byte(`{"collections":[{"address":"0xaaa","name":"alpha","description":"first","moderatorCount":2,"nextTokenId":7}]}`),
			}, nil)
		}

		t.Run("Should decode the collections payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingAgentCollectionsOverHttpBeforeEach(t)
			initSuite(suite)

			payload, err := suite.sut.List(context.Background())

			assert.NoError(t, err)
			if assert.NotNil(t, payload) {
				assert.Len(t, payload.Collections, 1)
				assert.Equal(t, "0xaaa", payload.Collections[0].Address)
				assert.Equal(t, "alpha", payload.Collections[0].Name)
				assert.Equal(t, uint64(2), payload.Collections[0].ModeratorCount)
			}
		})
	})

	t.Run("Given the runtime fails to read the chain", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingAgentCollectionsOverHttpTestingSuite) {
			suite.mockClient.EXPECT().GetInternalAgentCollectionsWithResponse(gomock.Any()).Return(&clients.GetInternalAgentCollectionsResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusBadGateway},
				Body:         []byte(`{"code":"collections_read_failed","message":"rpc down"}`),
			}, nil)
		}

		t.Run("Should return an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingAgentCollectionsOverHttpBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.List(context.Background())

			var target *apperrors.UnavailableError
			assert.ErrorAs(t, err, &target)
		})
	})
}
