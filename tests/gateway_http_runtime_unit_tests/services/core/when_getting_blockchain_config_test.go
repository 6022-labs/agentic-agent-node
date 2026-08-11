package services_test

import (
	"context"
	"net/http"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	services "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services/core"
	clients_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_http_runtime_mocks/clients_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenGettingBlockchainConfigTestingSuite struct {
	sut *services.HttpBlockchainConfigClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenGettingBlockchainConfigBeforeEach(t *testing.T) *WhenGettingBlockchainConfigTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenGettingBlockchainConfigTestingSuite{
		sut:        services.NewHttpBlockchainConfigClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenGettingBlockchainConfig(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime returns the chain config", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingBlockchainConfigTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalConfigBlockchainWithResponse(gomock.Any()).
				Return(&clients.GetInternalConfigBlockchainResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusOK},
					Body: []byte(
						`{"evmChains":{"80002":{"collectionsManagerAddress":"0xmanager","collectionCreatorAddress":"0xcreator","token6022Address":"0xtoken","ensRegistryAddress":"0xens","mintable":true,"facilitatorNetwork":"eip155:80002"}}}`,
					),
				}, nil)
		}

		t.Run("Should decode the config payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingBlockchainConfigBeforeEach(t)
			initSuite(suite)

			payload, err := suite.sut.Get(context.Background())

			assert.NoError(t, err)
			if assert.NotNil(t, payload) {
				if assert.Len(t, payload.EvmChains, 1) {
					chain := payload.EvmChains["80002"]
					assert.Equal(t, "0xmanager", chain.CollectionsManagerAddress)
					assert.Equal(t, "0xcreator", chain.CollectionCreatorAddress)
					assert.Equal(t, "0xtoken", chain.Token6022Address)
					assert.Equal(t, "0xens", chain.EnsRegistryAddress)
					assert.True(t, chain.Mintable)
					if assert.NotNil(t, chain.FacilitatorNetwork) {
						assert.Equal(t, "eip155:80002", *chain.FacilitatorNetwork)
					}
				}
			}
		})
	})

	t.Run("Given the runtime is unreachable", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingBlockchainConfigTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalConfigBlockchainWithResponse(gomock.Any()).
				Return(&clients.GetInternalConfigBlockchainResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError},
					Body:         []byte(`{"code":"boom"}`),
				}, nil)
		}

		t.Run("Should return an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingBlockchainConfigBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Get(context.Background())

			var target *apperrors.UnavailableError
			assert.ErrorAs(t, err, &target)
		})
	})
}
