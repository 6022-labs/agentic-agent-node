package services_test

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment_blockchain/services"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_blockchain_mocks/abi_mocks/interfaces_mocks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

const (
	erc20TestAsset   = "0x41E94Eb019C0762f9Bfcf9Fb1E58725BfB0e7582"
	erc20TestChainId = int64(80002)
)

type WhenValidatingErc20TestingSuite struct {
	sut *services.Erc20Validator

	erc20CallerMock *interfaces_mocks.MockErc20CallerInterface
}

func WhenValidatingErc20BeforeEach(t *testing.T) *WhenValidatingErc20TestingSuite {
	mockController := gomock.NewController(t)

	erc20CallerMock := interfaces_mocks.NewMockErc20CallerInterface(mockController)

	return &WhenValidatingErc20TestingSuite{
		sut:             services.NewErc20Validator(erc20CallerMock),
		erc20CallerMock: erc20CallerMock,
	}
}

func TestWhenValidatingErc20(t *testing.T) {
	t.Parallel()

	t.Run("Given the token exposes eip712Domain", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenValidatingErc20TestingSuite) {
			calls := erc20Probes(t, suite)
			calls = append(calls, suite.erc20CallerMock.EXPECT().
				CallContract(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(packErc20Output(t, "eip712Domain",
					[1]byte{0x0f}, "USD Coin", "2", big.NewInt(erc20TestChainId),
					common.HexToAddress(erc20TestAsset), [32]byte{}, []*big.Int{}), nil))
			gomock.InOrder(calls...)
		}

		t.Run("Should detect eip3009 and return the EIP-712 domain", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingErc20BeforeEach(t)
			initSuite(suite)

			method, name, version, err := suite.sut.ValidateErc20(context.Background(), erc20TestChainId, erc20TestAsset)

			assert.NoError(t, err)
			assert.Equal(t, enums.AssetTransferMethodEip3009, method)
			assert.Equal(t, "USD Coin", name)
			assert.Equal(t, "2", version)
		})
	})

	t.Run("Given no eip712Domain but name/version match DOMAIN_SEPARATOR", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenValidatingErc20TestingSuite) {
			calls := erc20Probes(t, suite)
			calls = append(calls,
				suite.erc20CallerMock.EXPECT().CallContract(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("execution reverted")), // eip712Domain()
				suite.erc20CallerMock.EXPECT().CallContract(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(packErc20Output(t, "name", "USD Coin"), nil),
				suite.erc20CallerMock.EXPECT().CallContract(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(packErc20Output(t, "version", "2"), nil),
				suite.erc20CallerMock.EXPECT().CallContract(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(packErc20Output(t, "DOMAIN_SEPARATOR", eip712DomainSeparator(t, "USD Coin", "2")), nil),
			)
			gomock.InOrder(calls...)
		}

		t.Run("Should detect eip3009 via the fallback", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingErc20BeforeEach(t)
			initSuite(suite)

			method, name, version, err := suite.sut.ValidateErc20(context.Background(), erc20TestChainId, erc20TestAsset)

			assert.NoError(t, err)
			assert.Equal(t, enums.AssetTransferMethodEip3009, method)
			assert.Equal(t, "USD Coin", name)
			assert.Equal(t, "2", version)
		})
	})

	t.Run("Given the token has no usable EIP-712 domain", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenValidatingErc20TestingSuite) {
			calls := erc20Probes(t, suite)
			calls = append(calls,
				suite.erc20CallerMock.EXPECT().CallContract(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("execution reverted")), // eip712Domain()
				suite.erc20CallerMock.EXPECT().CallContract(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(packErc20Output(t, "name", "Some Token"), nil),
				suite.erc20CallerMock.EXPECT().CallContract(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(packErc20Output(t, "version", "1"), nil),
				suite.erc20CallerMock.EXPECT().CallContract(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(packErc20Output(t, "DOMAIN_SEPARATOR", [32]byte{0x01}), nil), // mismatch
			)
			gomock.InOrder(calls...)
		}

		t.Run("Should fall back to permit2 with no domain", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingErc20BeforeEach(t)
			initSuite(suite)

			method, name, version, err := suite.sut.ValidateErc20(context.Background(), erc20TestChainId, erc20TestAsset)

			assert.NoError(t, err)
			assert.Equal(t, enums.AssetTransferMethodPermit2, method)
			assert.Empty(t, name)
			assert.Empty(t, version)
		})
	})

	t.Run("Given no eip712Domain and name() reverts", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenValidatingErc20TestingSuite) {
			calls := erc20Probes(t, suite)
			calls = append(calls,
				suite.erc20CallerMock.EXPECT().CallContract(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("execution reverted")), // eip712Domain()
				suite.erc20CallerMock.EXPECT().CallContract(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("execution reverted")), // name()
				suite.erc20CallerMock.EXPECT().CallContract(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(packErc20Output(t, "version", "2"), nil),
			)
			gomock.InOrder(calls...)
		}

		t.Run("Should fall back to permit2 with no domain", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingErc20BeforeEach(t)
			initSuite(suite)

			method, name, version, err := suite.sut.ValidateErc20(context.Background(), erc20TestChainId, erc20TestAsset)

			assert.NoError(t, err)
			assert.Equal(t, enums.AssetTransferMethodPermit2, method)
			assert.Empty(t, name)
			assert.Empty(t, version)
		})
	})

	t.Run("Given no eip712Domain and DOMAIN_SEPARATOR() reverts", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenValidatingErc20TestingSuite) {
			calls := erc20Probes(t, suite)
			calls = append(calls,
				suite.erc20CallerMock.EXPECT().CallContract(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("execution reverted")), // eip712Domain()
				suite.erc20CallerMock.EXPECT().CallContract(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(packErc20Output(t, "name", "USD Coin"), nil),
				suite.erc20CallerMock.EXPECT().CallContract(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(packErc20Output(t, "version", "2"), nil),
				suite.erc20CallerMock.EXPECT().CallContract(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("execution reverted")), // DOMAIN_SEPARATOR()
			)
			gomock.InOrder(calls...)
		}

		t.Run("Should fall back to permit2 with no domain", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingErc20BeforeEach(t)
			initSuite(suite)

			method, name, version, err := suite.sut.ValidateErc20(context.Background(), erc20TestChainId, erc20TestAsset)

			assert.NoError(t, err)
			assert.Equal(t, enums.AssetTransferMethodPermit2, method)
			assert.Empty(t, name)
			assert.Empty(t, version)
		})
	})

	t.Run("Given the asset is not an ERC-20", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenValidatingErc20TestingSuite) {
			suite.erc20CallerMock.EXPECT().
				CallContract(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, errors.New("execution reverted")).
				Times(1)
		}

		t.Run("Should reject the asset", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingErc20BeforeEach(t)
			initSuite(suite)

			_, _, _, err := suite.sut.ValidateErc20(context.Background(), erc20TestChainId, erc20TestAsset)

			assert.ErrorContains(t, err, "no ERC-20 token answers decimals()")
		})
	})

	t.Run("Given the address is not a contract", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenValidatingErc20TestingSuite) {
			// An EOA answers eth_call with empty data; the binding then checks
			// the code at the address and finds none.
			suite.erc20CallerMock.EXPECT().
				CallContract(gomock.Any(), gomock.Any(), gomock.Any()).
				Return([]byte{}, nil).
				Times(1)
			suite.erc20CallerMock.EXPECT().
				CodeAt(gomock.Any(), gomock.Any(), gomock.Any()).
				Return([]byte{}, nil).
				Times(1)
		}

		t.Run("Should reject the address as not deployed", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingErc20BeforeEach(t)
			initSuite(suite)

			_, _, _, err := suite.sut.ValidateErc20(context.Background(), erc20TestChainId, erc20TestAsset)

			assert.ErrorContains(t, err, "no contract deployed")
		})
	})
}
