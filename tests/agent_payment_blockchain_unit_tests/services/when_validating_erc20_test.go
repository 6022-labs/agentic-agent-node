package services_test

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"testing"

	contract_abi "github.com/6022-labs/agentic-agent-node/src/agent_payment_blockchain/abi"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment_blockchain/services"
	runtime_settings "github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_blockchain_mocks/abi_mocks/interfaces_mocks"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

const erc20TestAsset = "0x41E94Eb019C0762f9Bfcf9Fb1E58725BfB0e7582"

type WhenValidatingErc20TestingSuite struct {
	sut *services.Erc20Validator

	erc20CallerMock *interfaces_mocks.MockErc20CallerInterface
}

func WhenValidatingErc20BeforeEach(t *testing.T) *WhenValidatingErc20TestingSuite {
	mockController := gomock.NewController(t)

	erc20CallerMock := interfaces_mocks.NewMockErc20CallerInterface(mockController)

	return &WhenValidatingErc20TestingSuite{
		sut: services.NewErc20Validator(
			erc20CallerMock,
			&runtime_settings.BlockchainSettings{Id: 80002},
		),

		erc20CallerMock: erc20CallerMock,
	}
}

// packErc20Output encodes a method return the way the contract would.
func packErc20Output(t *testing.T, method string, values ...any) []byte {
	t.Helper()

	parsed, err := abi.JSON(strings.NewReader(contract_abi.Erc20MetaData.ABI))
	assert.NoError(t, err)
	packed, err := parsed.Methods[method].Outputs.Pack(values...)
	assert.NoError(t, err)
	return packed
}

func TestWhenValidatingErc20(t *testing.T) {
	t.Parallel()

	t.Run("Given the asset answers the ERC-20 probes", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenValidatingErc20TestingSuite) {
			gomock.InOrder(
				suite.erc20CallerMock.EXPECT().
					CallContract(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(packErc20Output(t, "decimals", uint8(6)), nil),
				suite.erc20CallerMock.EXPECT().
					CallContract(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(packErc20Output(t, "symbol", "USDC"), nil),
				suite.erc20CallerMock.EXPECT().
					CallContract(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(packErc20Output(t, "totalSupply", big.NewInt(1_000_000)), nil),
			)
		}

		t.Run("Should accept the asset", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingErc20BeforeEach(t)
			initSuite(suite)

			err := suite.sut.ValidateErc20(context.Background(), 80002, erc20TestAsset)

			assert.NoError(t, err)
		})
	})

	t.Run("Given the option targets another chain", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenValidatingErc20TestingSuite) {
			suite.erc20CallerMock.EXPECT().
				CallContract(gomock.Any(), gomock.Any(), gomock.Any()).
				Times(0)
		}

		t.Run("Should skip validation (this node has no RPC for it)", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingErc20BeforeEach(t)
			initSuite(suite)

			err := suite.sut.ValidateErc20(context.Background(), 137, erc20TestAsset)

			assert.NoError(t, err)
		})
	})

	t.Run("Given the contract call reverts", func(t *testing.T) {
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

			err := suite.sut.ValidateErc20(context.Background(), 80002, erc20TestAsset)

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

			err := suite.sut.ValidateErc20(context.Background(), 80002, erc20TestAsset)

			assert.ErrorContains(t, err, "no contract deployed")
		})
	})

	t.Run("Given a contract that answers with non-ERC-20 data", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenValidatingErc20TestingSuite) {
			suite.erc20CallerMock.EXPECT().
				CallContract(gomock.Any(), gomock.Any(), gomock.Any()).
				Return([]byte{0x01, 0x02, 0x03}, nil).
				Times(1)
		}

		t.Run("Should reject the asset on decode failure", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingErc20BeforeEach(t)
			initSuite(suite)

			err := suite.sut.ValidateErc20(context.Background(), 80002, erc20TestAsset)

			assert.ErrorContains(t, err, "no ERC-20 token answers decimals()")
		})
	})
}
