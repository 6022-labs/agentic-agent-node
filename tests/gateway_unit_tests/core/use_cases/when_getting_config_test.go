package use_cases_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/models"
	core_settings "github.com/6022-labs/agentic-agent-node/src/gateway/core/settings"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/use_cases"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/core_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenGettingConfigTestingSuite struct {
	sut *use_cases.GetConfig

	mockClient *interfaces_mocks.MockBlockchainConfigClientInterface
}

func WhenGettingConfigBeforeEach(t *testing.T) *WhenGettingConfigTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := interfaces_mocks.NewMockBlockchainConfigClientInterface(mockController)

	ipfsGatewaysSettings := &core_settings.IPFSGatewaysSettings{
		GatewayUrlTemplates: []string{"https://{cid}.ipfs.example.com"},
	}

	sut := use_cases.NewGetConfig(mockClient, ipfsGatewaysSettings)

	return &WhenGettingConfigTestingSuite{
		sut: sut,

		mockClient: mockClient,
	}
}

func TestWhenGettingConfig(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingConfigTestingSuite) {
			suite.mockClient.EXPECT().Get(gomock.Any()).
				Return(nil, errors.NewUnavailableError("runtime_unavailable", "down"))
		}

		t.Run("Should surface the UnavailableError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingConfigBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background())

			var target *errors.UnavailableError
			assert.ErrorAs(t, err, &target)
		})
	})

	t.Run("Given the runtime returns the chain config", func(t *testing.T) {
		t.Parallel()

		facilitatorNetwork := "eip155:80002"
		initSuite := func(suite *WhenGettingConfigTestingSuite) {
			suite.mockClient.EXPECT().Get(gomock.Any()).Return(&models.ExternalBlockchainConfigPayload{
				EvmChains: map[string]models.ExternalEvmChainConfig{
					"80002": {
						CollectionsManagerAddress: "0xmanager",
						CollectionCreatorAddress:  "0xcreator",
						Token6022Address:          "0xtoken",
						EnsRegistryAddress:        "0xens",
						Mintable:                  true,
						FacilitatorNetwork:        &facilitatorNetwork,
					},
				},
			}, nil)
		}

		t.Run("Should map every field", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingConfigBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			if assert.Len(t, resp.EvmChains, 1) {
				chain := resp.EvmChains["80002"]
				assert.Equal(t, "0xmanager", chain.CollectionsManagerAddress)
				assert.Equal(t, "0xcreator", chain.CollectionCreatorAddress)
				assert.Equal(t, "0xtoken", chain.Token6022Address)
				assert.Equal(t, "0xens", chain.EnsRegistryAddress)
				assert.True(t, chain.Mintable)
				if assert.NotNil(t, chain.FacilitatorNetwork) {
					assert.Equal(t, "eip155:80002", *chain.FacilitatorNetwork)
				}
			}
			assert.Equal(t, []string{"https://{cid}.ipfs.example.com"}, resp.IpfsGatewayUrlTemplates)
		})
	})
}
