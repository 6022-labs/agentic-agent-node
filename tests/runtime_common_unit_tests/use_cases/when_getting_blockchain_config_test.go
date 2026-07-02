package use_cases_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	"github.com/6022-labs/agentic-agent-node/src/runtime_common/use_cases"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

type WhenGettingBlockchainConfigTestingSuite struct {
	sut *use_cases.GetBlockchainConfig
}

func WhenGettingBlockchainConfigBeforeEach(_ *testing.T) *WhenGettingBlockchainConfigTestingSuite {
	chains := settings.EvmChainsSettings{
		"80002": {ChainId: "80002", HttpUrl: "http://localhost:8545"},
	}
	managers := settings.AgentCollectionsManagersSettings{
		"80002": common.HexToAddress("0x2e0A2C00601F7DCa8B4b7ae05A5fd27bb46080d3"),
	}
	creators := settings.AgentCollectionCreatorsSettings{
		"80002": common.HexToAddress("0x9D72199fC7C00416ceD5bC157Bdb2e3775511717"),
	}
	token6022 := settings.Token6022Settings{
		"80002": common.HexToAddress("0xbB2CEbf252EaDdf1674Cc7C458FCf4877704820F"),
	}
	ensRegistries := settings.AgentEnsRegistriesSettings{
		"80002": common.HexToAddress("0xb02c3D72cf0939c38C26291e5645d25De18478b7"),
	}

	sut := use_cases.NewGetBlockchainConfig(
		&chains,
		&managers,
		&creators,
		&token6022,
		&ensRegistries,
		&settings.PaymentFacilitatorsSettings{
			"eip155:80002": {BaseUrl: "https://x402-amoy.polygon.technology"},
		},
	)

	return &WhenGettingBlockchainConfigTestingSuite{sut: sut}
}

func TestWhenGettingBlockchainConfig(t *testing.T) {
	t.Parallel()

	t.Run("Given a reachable chain", func(t *testing.T) {
		t.Parallel()

		t.Run("Should key the chain by its id with every contract address", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingBlockchainConfigBeforeEach(t)

			resp, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			if assert.Len(t, resp.Chains, 1) {
				chain := resp.Chains["80002"]
				assert.Equal(t, "0x2e0A2C00601F7DCa8B4b7ae05A5fd27bb46080d3", chain.CollectionsManagerAddress)
				assert.Equal(t, "0x9D72199fC7C00416ceD5bC157Bdb2e3775511717", chain.CollectionCreatorAddress)
				assert.Equal(t, "0xbB2CEbf252EaDdf1674Cc7C458FCf4877704820F", chain.Token6022Address)
				assert.Equal(t, "0xb02c3D72cf0939c38C26291e5645d25De18478b7", chain.EnsRegistryAddress)
				assert.True(t, chain.Mintable)
				if assert.NotNil(t, chain.FacilitatorNetwork) {
					assert.Equal(t, "eip155:80002", *chain.FacilitatorNetwork)
				}
			}
		})
	})
}
