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
	sut := use_cases.NewGetBlockchainConfig(
		&settings.BlockchainSettings{Id: 80002},
		&settings.AgentCollectionsManagersSettings{
			ManagerAddress: common.HexToAddress("0x2e0A2C00601F7DCa8B4b7ae05A5fd27bb46080d3"),
		},
		&settings.AgentCollectionCreatorsSettings{
			CreatorAddress: common.HexToAddress("0x9D72199fC7C00416ceD5bC157Bdb2e3775511717"),
		},
		&settings.Token6022Settings{
			SmartContractAddress: common.HexToAddress("0xbB2CEbf252EaDdf1674Cc7C458FCf4877704820F"),
		},
		&settings.AgentEnsRegistriesSettings{
			RegistryAddress: common.HexToAddress("0xb02c3D72cf0939c38C26291e5645d25De18478b7"),
		},
	)

	return &WhenGettingBlockchainConfigTestingSuite{sut: sut}
}

func TestWhenGettingBlockchainConfig(t *testing.T) {
	t.Parallel()

	t.Run("Given the node was bound to a chain at startup", func(t *testing.T) {
		t.Parallel()

		t.Run("Should compose the chain id and every contract address", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingBlockchainConfigBeforeEach(t)

			resp, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, uint64(80002), resp.ChainId)
			assert.Equal(t, "0x2e0A2C00601F7DCa8B4b7ae05A5fd27bb46080d3", resp.CollectionsManagerAddress)
			assert.Equal(t, "0x9D72199fC7C00416ceD5bC157Bdb2e3775511717", resp.CollectionCreatorAddress)
			assert.Equal(t, "0xbB2CEbf252EaDdf1674Cc7C458FCf4877704820F", resp.Token6022Address)
			assert.Equal(t, "0xb02c3D72cf0939c38C26291e5645d25De18478b7", resp.EnsRegistryAddress)
		})
	})
}
