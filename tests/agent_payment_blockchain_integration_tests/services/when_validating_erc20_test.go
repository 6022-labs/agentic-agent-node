package services_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	payment_factory "github.com/6022-labs/agentic-agent-node/src/agent_payment_blockchain/factory"
	payment_services "github.com/6022-labs/agentic-agent-node/src/agent_payment_blockchain/services"
	"github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	"github.com/6022-labs/agentic-agent-node/tests/agent_payment_blockchain_integration_tests/fixtures"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/container_builders"
	runtime_metrics_mocks "github.com/6022-labs/agentic-agent-node/tests/runtime_common_mocks/metrics_mocks/interfaces_mocks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenValidatingErc20TestingSuite struct {
	anvilAccountSettings container_builders.AnvilAccountSettings
	client               *ethclient.Client
	sut                  *payment_services.Erc20Validator
}

func WhenValidatingErc20BeforeEach(t *testing.T) *WhenValidatingErc20TestingSuite {
	anvilAccountSettings := container_builders.GetAnvilContainerBuilder().Create()

	client, err := ethclient.Dial(anvilAccountSettings.HttpUrl)
	if err != nil {
		t.Fatalf("failed to dial anvil: %v", err)
	}
	t.Cleanup(client.Close)

	mockBlockchainRpcMetrics := runtime_metrics_mocks.NewMockBlockchainRpcMetricsInterface(gomock.NewController(t))
	mockBlockchainRpcMetrics.EXPECT().RecordRequest(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	return &WhenValidatingErc20TestingSuite{
		anvilAccountSettings: anvilAccountSettings,
		client:               client,
		sut: payment_services.NewErc20Validator(
			payment_factory.NewEthClientFactory(
				evmChainsSettings(anvilAccountSettings),
				mockBlockchainRpcMetrics,
			),
		),
	}
}

func TestWhenValidatingErc20(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given a token exposing an ERC-5267 EIP-712 domain", func(t *testing.T) {
		t.Parallel()

		t.Run("Should accept it as an EIP-3009 asset carrying the domain name", func(t *testing.T) {
			t.Parallel()
			suite := WhenValidatingErc20BeforeEach(t)
			asset, err := fixtures.DeployErc20(
				ctx,
				suite.client,
				suite.anvilAccountSettings,
				fixtures.Erc20WithDomain,
				"Sixty Twenty Two",
				"6022",
			)
			assert.NoError(t, err)

			method, name, version, err := suite.sut.ValidateErc20(ctx, suite.anvilAccountSettings.ChainId, asset.Hex())

			assert.NoError(t, err)
			assert.Equal(t, enums.AssetTransferMethodEip3009, method)
			assert.Equal(t, "Sixty Twenty Two", name)
			assert.Equal(t, "1", version)
		})
	})

	t.Run("Given a plain ERC-20 with no EIP-712 domain", func(t *testing.T) {
		t.Parallel()

		t.Run("Should accept it but route it through permit2", func(t *testing.T) {
			t.Parallel()
			suite := WhenValidatingErc20BeforeEach(t)
			asset, err := fixtures.DeployErc20(
				ctx,
				suite.client,
				suite.anvilAccountSettings,
				fixtures.Erc20WithoutDomain,
				"Plain Token",
				"PLAIN",
			)
			assert.NoError(t, err)

			method, name, version, err := suite.sut.ValidateErc20(ctx, suite.anvilAccountSettings.ChainId, asset.Hex())

			assert.NoError(t, err)
			assert.Equal(t, enums.AssetTransferMethodPermit2, method)
			assert.Empty(t, name)
			assert.Empty(t, version)
		})
	})

	t.Run("Given an address with no contract code", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject it", func(t *testing.T) {
			t.Parallel()
			suite := WhenValidatingErc20BeforeEach(t)

			_, _, _, err := suite.sut.ValidateErc20(
				ctx,
				suite.anvilAccountSettings.ChainId,
				common.HexToAddress("0x000000000000000000000000000000000000dEaD").Hex(),
			)

			assert.Error(t, err)
		})
	})

	t.Run("Given an address holding code that is not an ERC-20", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject it", func(t *testing.T) {
			t.Parallel()
			suite := WhenValidatingErc20BeforeEach(t)
			asset := container_builders.GetAnvilContainerBuilder().CreateContractlessAddress()

			_, _, _, err := suite.sut.ValidateErc20(ctx, suite.anvilAccountSettings.ChainId, asset.Hex())

			assert.Error(t, err)
		})
	})
}

// evmChainsSettings points the eth client factory at the test's anvil chain.
func evmChainsSettings(anvilAccountSettings container_builders.AnvilAccountSettings) *settings.EvmChainsSettings {
	evmChains := settings.EvmChainsSettings{
		anvilAccountSettings.ChainId: {
			ChainId: anvilAccountSettings.ChainId,
			HttpUrl: anvilAccountSettings.HttpUrl,
		},
	}
	return &evmChains
}
