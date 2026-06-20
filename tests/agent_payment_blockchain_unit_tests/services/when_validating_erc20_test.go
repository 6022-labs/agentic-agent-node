package services_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment_blockchain/factory"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment_blockchain/services"
	"github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	metrics_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/runtime_common_mocks/metrics_mocks/interfaces_mocks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

const (
	erc20TestAsset      = "0x41E94Eb019C0762f9Bfcf9Fb1E58725BfB0e7582"
	erc20TestChainId    = int64(80002)
	erc20TestChainIdStr = "80002"
)

type WhenValidatingErc20TestingSuite struct {
	sut *services.Erc20Validator
}

// WhenValidatingErc20BeforeEach points the validator at a fake RPC scripted with
// the given per-method answers (hasCode backs the binding's no-code check).
func WhenValidatingErc20BeforeEach(t *testing.T, methods map[string]erc20Method, hasCode bool) *WhenValidatingErc20TestingSuite {
	server := newErc20RpcServer(t, methods, hasCode)
	t.Cleanup(server.Close)

	chains := settings.ChainsSettings{
		erc20TestChainIdStr: {ChainId: erc20TestChainIdStr, HttpUrl: server.URL},
	}

	mockMetrics := metrics_interfaces_mocks.NewMockBlockchainRpcMetricsInterface(gomock.NewController(t))
	mockMetrics.EXPECT().RecordRequest(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	return &WhenValidatingErc20TestingSuite{
		sut: services.NewErc20Validator(factory.NewEthClientFactory(&chains, mockMetrics)),
	}
}

func TestWhenValidatingErc20(t *testing.T) {
	t.Parallel()

	t.Run("Given the token exposes eip712Domain", func(t *testing.T) {
		t.Parallel()

		methods := erc20ProbeMethods(t)
		methods["eip712Domain"] = erc20Method{data: packErc20Output(t, "eip712Domain",
			[1]byte{0x0f}, "USD Coin", "2", big.NewInt(erc20TestChainId),
			common.HexToAddress(erc20TestAsset), [32]byte{}, []*big.Int{})}

		t.Run("Should detect eip3009 and return the EIP-712 domain", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingErc20BeforeEach(t, methods, true)

			method, name, version, err := suite.sut.ValidateErc20(context.Background(), erc20TestChainIdStr, erc20TestAsset)

			assert.NoError(t, err)
			assert.Equal(t, enums.AssetTransferMethodEip3009, method)
			assert.Equal(t, "USD Coin", name)
			assert.Equal(t, "2", version)
		})
	})

	t.Run("Given no eip712Domain but name/version match DOMAIN_SEPARATOR", func(t *testing.T) {
		t.Parallel()

		methods := erc20ProbeMethods(t)
		methods["name"] = erc20Method{data: packErc20Output(t, "name", "USD Coin")}
		methods["version"] = erc20Method{data: packErc20Output(t, "version", "2")}
		methods["DOMAIN_SEPARATOR"] = erc20Method{data: packErc20Output(t, "DOMAIN_SEPARATOR", eip712DomainSeparator(t, "USD Coin", "2"))}

		t.Run("Should detect eip3009 via the fallback", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingErc20BeforeEach(t, methods, true)

			method, name, version, err := suite.sut.ValidateErc20(context.Background(), erc20TestChainIdStr, erc20TestAsset)

			assert.NoError(t, err)
			assert.Equal(t, enums.AssetTransferMethodEip3009, method)
			assert.Equal(t, "USD Coin", name)
			assert.Equal(t, "2", version)
		})
	})

	t.Run("Given the token has no usable EIP-712 domain", func(t *testing.T) {
		t.Parallel()

		methods := erc20ProbeMethods(t)
		methods["name"] = erc20Method{data: packErc20Output(t, "name", "Some Token")}
		methods["version"] = erc20Method{data: packErc20Output(t, "version", "1")}
		methods["DOMAIN_SEPARATOR"] = erc20Method{data: packErc20Output(t, "DOMAIN_SEPARATOR", [32]byte{0x01})}

		t.Run("Should fall back to permit2 with no domain", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingErc20BeforeEach(t, methods, true)

			method, name, version, err := suite.sut.ValidateErc20(context.Background(), erc20TestChainIdStr, erc20TestAsset)

			assert.NoError(t, err)
			assert.Equal(t, enums.AssetTransferMethodPermit2, method)
			assert.Empty(t, name)
			assert.Empty(t, version)
		})
	})

	t.Run("Given no eip712Domain and name() reverts", func(t *testing.T) {
		t.Parallel()

		methods := erc20ProbeMethods(t)
		methods["version"] = erc20Method{data: packErc20Output(t, "version", "2")}

		t.Run("Should fall back to permit2 with no domain", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingErc20BeforeEach(t, methods, true)

			method, name, version, err := suite.sut.ValidateErc20(context.Background(), erc20TestChainIdStr, erc20TestAsset)

			assert.NoError(t, err)
			assert.Equal(t, enums.AssetTransferMethodPermit2, method)
			assert.Empty(t, name)
			assert.Empty(t, version)
		})
	})

	t.Run("Given no eip712Domain and DOMAIN_SEPARATOR() reverts", func(t *testing.T) {
		t.Parallel()

		methods := erc20ProbeMethods(t)
		methods["name"] = erc20Method{data: packErc20Output(t, "name", "USD Coin")}
		methods["version"] = erc20Method{data: packErc20Output(t, "version", "2")}

		t.Run("Should fall back to permit2 with no domain", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingErc20BeforeEach(t, methods, true)

			method, name, version, err := suite.sut.ValidateErc20(context.Background(), erc20TestChainIdStr, erc20TestAsset)

			assert.NoError(t, err)
			assert.Equal(t, enums.AssetTransferMethodPermit2, method)
			assert.Empty(t, name)
			assert.Empty(t, version)
		})
	})

	t.Run("Given the asset is not an ERC-20", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject the asset", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingErc20BeforeEach(t, map[string]erc20Method{}, true)

			_, _, _, err := suite.sut.ValidateErc20(context.Background(), erc20TestChainIdStr, erc20TestAsset)

			assert.ErrorContains(t, err, "no ERC-20 token answers decimals()")
		})
	})

	t.Run("Given the address is not a contract", func(t *testing.T) {
		t.Parallel()

		// An EOA answers eth_call with empty data; the binding then checks the
		// code at the address and finds none.
		methods := map[string]erc20Method{"decimals": {empty: true}}

		t.Run("Should reject the address as not deployed", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingErc20BeforeEach(t, methods, false)

			_, _, _, err := suite.sut.ValidateErc20(context.Background(), erc20TestChainIdStr, erc20TestAsset)

			assert.ErrorContains(t, err, "no contract deployed")
		})
	})
}
