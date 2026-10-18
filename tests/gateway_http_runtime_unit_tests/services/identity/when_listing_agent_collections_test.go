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

type WhenListingAgentCollectionsTestingSuite struct {
	sut *services.HttpAgentCollectionsClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenListingAgentCollectionsBeforeEach(t *testing.T) *WhenListingAgentCollectionsTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenListingAgentCollectionsTestingSuite{
		sut:        services.NewHttpAgentCollectionsClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenListingAgentCollections(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingAgentCollectionsTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalAgentCollectionsWithResponse(gomock.Any(), gomock.Any()).
				Return(nil, errors.New("runtime down"))
		}

		t.Run("Should map the transport error onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingAgentCollectionsBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.List(context.Background(), "80002")

			var target *apperrors.UnavailableError
			if assert.ErrorAs(t, err, &target) {
				assert.Equal(t, "runtime_unavailable", target.Code)
			}
		})
	})

	t.Run("Given the runtime answers 400", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingAgentCollectionsTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalAgentCollectionsWithResponse(gomock.Any(), gomock.Any()).
				Return(&clients.GetInternalAgentCollectionsResponse{HTTPResponse: &http.Response{StatusCode: http.StatusBadRequest}, Body: []byte(`{"code":"x","field":"f","message":"m"}`)}, nil)
		}

		t.Run("Should map the 400 onto a ValidationError", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingAgentCollectionsBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.List(context.Background(), "80002")

			var target *apperrors.ValidationError
			assert.ErrorAs(t, err, &target)
		})
	})

	t.Run("Given the runtime answers 502", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingAgentCollectionsTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalAgentCollectionsWithResponse(gomock.Any(), &clients.GetInternalAgentCollectionsParams{ChainId: "80002"}).
				Return(&clients.GetInternalAgentCollectionsResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusBadGateway},
					Body:         []byte(`{"code":"collections_read_failed","message":"rpc down"}`),
				}, nil)
		}

		t.Run("Should return an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingAgentCollectionsBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.List(context.Background(), "80002")

			var target *apperrors.UnavailableError
			assert.ErrorAs(t, err, &target)
		})
	})

	t.Run("Given the runtime returns the collections catalog", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingAgentCollectionsTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalAgentCollectionsWithResponse(gomock.Any(), &clients.GetInternalAgentCollectionsParams{ChainId: "80002"}).
				Return(&clients.GetInternalAgentCollectionsResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusOK},
					Body: []byte(
						`{"collections":[{"address":"0xaaa","name":"alpha","description":"first","moderatorCount":2,"nextTokenId":7}]}`,
					),
				}, nil)
		}

		t.Run("Should decode the collections payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingAgentCollectionsBeforeEach(t)
			initSuite(suite)

			payload, err := suite.sut.List(context.Background(), "80002")

			assert.NoError(t, err)
			if assert.NotNil(t, payload) {
				assert.Len(t, payload.Collections, 1)
				assert.Equal(t, "0xaaa", payload.Collections[0].Address)
				assert.Equal(t, "alpha", payload.Collections[0].Name)
				assert.Equal(t, uint64(2), payload.Collections[0].ModeratorCount)
			}
		})
	})
}
