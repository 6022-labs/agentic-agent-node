package services_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment_blockchain/services"
	"github.com/6022-labs/agentic-agent-node/src/common/x402"
	"github.com/6022-labs/agentic-agent-node/tests/agent_payment_blockchain_mocks/factory_mocks/interfaces_mocks"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

const (
	permit2Network    = "eip155:80002"
	permit2ChainId    = "80002"
	permit2EvmChainId = 80002
	testErc20Asset    = "0x41E94Eb019C0762f9Bfcf9Fb1E58725BfB0e7582"
)

type WhenEnsuringPermit2ApprovalTestingSuite struct {
	sut *services.Permit2ApprovalEnsurer

	mockEthClientFactory  *interfaces_mocks.MockEthClientFactoryInterface
	mockTransactorFactory *interfaces_mocks.MockTransactorFactoryInterface
	submitted             *[]approveCall
}

// WhenEnsuringPermit2ApprovalBeforeEach records every approve; successive allowances model a changing value.
func WhenEnsuringPermit2ApprovalBeforeEach(
	t *testing.T,
	allowances ...*big.Int,
) *WhenEnsuringPermit2ApprovalTestingSuite {
	t.Helper()

	mockController := gomock.NewController(t)

	server, submitted := approvalRpcServer(t, allowances)
	client, err := ethclient.Dial(server.URL)
	assert.NoError(t, err)
	t.Cleanup(client.Close)

	mockEthClientFactory := interfaces_mocks.NewMockEthClientFactoryInterface(mockController)
	mockTransactorFactory := interfaces_mocks.NewMockTransactorFactoryInterface(mockController)
	mockEthClientFactory.EXPECT().For(permit2ChainId).Return(client, nil).AnyTimes()

	return &WhenEnsuringPermit2ApprovalTestingSuite{
		sut:                   services.NewPermit2ApprovalEnsurer(mockEthClientFactory, mockTransactorFactory),
		mockEthClientFactory:  mockEthClientFactory,
		mockTransactorFactory: mockTransactorFactory,
		submitted:             submitted,
	}
}

// expectTransactor hands back a real signer, so the approve path actually submits a transaction.
func (s *WhenEnsuringPermit2ApprovalTestingSuite) expectTransactor(t *testing.T) {
	t.Helper()

	key, err := crypto.GenerateKey()
	assert.NoError(t, err)
	transactor, err := bind.NewKeyedTransactorWithChainID(key, big.NewInt(permit2EvmChainId))
	assert.NoError(t, err)

	s.mockTransactorFactory.EXPECT().
		Create(gomock.Any(), uint64(permit2EvmChainId)).
		Return(transactor, nil).
		AnyTimes()
}

func (s *WhenEnsuringPermit2ApprovalTestingSuite) approvals() []approveCall {
	return *s.submitted
}

func TestWhenEnsuringPermit2Approval(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given the wallet already allows at least the amount", func(t *testing.T) {
		t.Parallel()

		t.Run("Should send no transaction", func(t *testing.T) {
			t.Parallel()

			suite := WhenEnsuringPermit2ApprovalBeforeEach(t, big.NewInt(50000))
			suite.expectTransactor(t)

			err := suite.sut.EnsureApproval(ctx, permit2Network, testErc20Asset, big.NewInt(10000))

			assert.NoError(t, err)
		})

		t.Run("Should treat an exactly-equal allowance as sufficient", func(t *testing.T) {
			t.Parallel()

			suite := WhenEnsuringPermit2ApprovalBeforeEach(t, big.NewInt(10000))
			suite.expectTransactor(t)

			err := suite.sut.EnsureApproval(ctx, permit2Network, testErc20Asset, big.NewInt(10000))

			assert.NoError(t, err)
		})
	})

	t.Run("Given a network that is not CAIP-2", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse before touching the chain", func(t *testing.T) {
			t.Parallel()

			suite := WhenEnsuringPermit2ApprovalBeforeEach(t, big.NewInt(0))
			suite.expectTransactor(t)

			err := suite.sut.EnsureApproval(ctx, "80002", testErc20Asset, big.NewInt(10000))

			assert.ErrorContains(t, err, "CAIP-2")
		})
	})

	t.Run("Given the wallet is not bootstrapped", func(t *testing.T) {
		t.Parallel()

		t.Run("Should propagate the failure rather than sign an unusable approval", func(t *testing.T) {
			t.Parallel()

			suite := WhenEnsuringPermit2ApprovalBeforeEach(t, big.NewInt(0))
			suite.mockTransactorFactory.EXPECT().
				Create(gomock.Any(), uint64(80002)).
				Return(nil, assert.AnError)

			err := suite.sut.EnsureApproval(ctx, permit2Network, testErc20Asset, big.NewInt(10000))

			assert.Error(t, err)
		})
	})

	t.Run("Given no existing allowance", func(t *testing.T) {
		t.Parallel()

		t.Run("Should approve Permit2 for the maximum, not just the amount", func(t *testing.T) {
			t.Parallel()

			suite := WhenEnsuringPermit2ApprovalBeforeEach(t, big.NewInt(0))
			suite.expectTransactor(t)

			err := suite.sut.EnsureApproval(ctx, permit2Network, testErc20Asset, big.NewInt(10000))

			assert.NoError(t, err)
			assert.Len(t, suite.approvals(), 1)
			assert.Equal(t, common.HexToAddress(x402.Permit2ContractAddress), suite.approvals()[0].spender)
			assert.Equal(t, maxUint256.String(), suite.approvals()[0].amount.String())
		})
	})

	t.Run("Given a non-zero allowance below the amount", func(t *testing.T) {
		t.Parallel()

		// Tokens like USDT revert on a non-zero to non-zero change.
		t.Run("Should clear it to zero before granting the maximum", func(t *testing.T) {
			t.Parallel()

			suite := WhenEnsuringPermit2ApprovalBeforeEach(t, big.NewInt(500))
			suite.expectTransactor(t)

			err := suite.sut.EnsureApproval(ctx, permit2Network, testErc20Asset, big.NewInt(10000))

			assert.NoError(t, err)
			assert.Len(t, suite.approvals(), 2)
			assert.Equal(t, "0", suite.approvals()[0].amount.String())
			assert.Equal(t, maxUint256.String(), suite.approvals()[1].amount.String())
		})
	})

	t.Run("Given an allowance already at or above the amount", func(t *testing.T) {
		t.Parallel()

		t.Run("Should send nothing", func(t *testing.T) {
			t.Parallel()

			suite := WhenEnsuringPermit2ApprovalBeforeEach(t, big.NewInt(10000))
			suite.expectTransactor(t)

			err := suite.sut.EnsureApproval(ctx, permit2Network, testErc20Asset, big.NewInt(10000))

			assert.NoError(t, err)
			assert.Empty(t, suite.approvals())
		})
	})

	t.Run("Given the canonical Permit2 address", func(t *testing.T) {
		t.Parallel()

		t.Run("Should be the same on every chain, as CREATE2 guarantees", func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, "0x000000000022D473030F116dDEE9F6B43aC78BA3", x402.Permit2ContractAddress)
		})
	})
}
