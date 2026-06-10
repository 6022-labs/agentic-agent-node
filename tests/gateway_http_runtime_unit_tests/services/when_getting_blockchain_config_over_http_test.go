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

type WhenGettingBlockchainConfigOverHttpTestingSuite struct {
	sut *services.HttpBlockchainConfigClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenGettingBlockchainConfigOverHttpBeforeEach(t *testing.T) *WhenGettingBlockchainConfigOverHttpTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenGettingBlockchainConfigOverHttpTestingSuite{
		sut:        services.NewHttpBlockchainConfigClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenGettingBlockchainConfigOverHttp(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime returns the chain config", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingBlockchainConfigOverHttpTestingSuite) {
			suite.mockClient.EXPECT().GetInternalConfigBlockchainWithResponse(gomock.Any()).Return(&clients.GetInternalConfigBlockchainResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusOK},
				Body:         []byte(`{"chainId":80002,"collectionsManagerAddress":"0xmanager","collectionCreatorAddress":"0xcreator","token6022Address":"0xtoken","ensRegistryAddress":"0xens"}`),
			}, nil)
		}

		t.Run("Should decode the config payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingBlockchainConfigOverHttpBeforeEach(t)
			initSuite(suite)

			payload, err := suite.sut.Get(context.Background())

			assert.NoError(t, err)
			if assert.NotNil(t, payload) {
				assert.Equal(t, uint64(80002), payload.ChainId)
				assert.Equal(t, "0xmanager", payload.CollectionsManagerAddress)
				assert.Equal(t, "0xcreator", payload.CollectionCreatorAddress)
				assert.Equal(t, "0xtoken", payload.Token6022Address)
				assert.Equal(t, "0xens", payload.EnsRegistryAddress)
			}
		})
	})

	t.Run("Given the runtime is unreachable", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingBlockchainConfigOverHttpTestingSuite) {
			suite.mockClient.EXPECT().GetInternalConfigBlockchainWithResponse(gomock.Any()).Return(&clients.GetInternalConfigBlockchainResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError},
				Body:         []byte(`{"code":"boom"}`),
			}, nil)
		}

		t.Run("Should return an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingBlockchainConfigOverHttpBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Get(context.Background())

			var target *apperrors.UnavailableError
			assert.ErrorAs(t, err, &target)
		})
	})
}
