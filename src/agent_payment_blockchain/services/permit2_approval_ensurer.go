package services

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment_blockchain/abi"
	factory_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment_blockchain/factory/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/common/helpers"
	"github.com/6022-labs/agentic-agent-node/src/common/x402"
	"github.com/6022-labs/agentic-agent-node/src/runtime_common/helper"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

// maxUint256 is granted once so a later payment pays no gas.
var maxUint256 = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))

// Permit2ApprovalEnsurer lets Permit2 pull the token at settle time; without it a signature is unsettleable.
type Permit2ApprovalEnsurer struct {
	ethClientFactory  factory_interfaces.EthClientFactoryInterface
	transactorFactory factory_interfaces.TransactorFactoryInterface

	// Concurrent turns would both read allowance 0 and reuse one wallet nonce, dropping an approval.
	approvalMutex sync.Mutex
}

func NewPermit2ApprovalEnsurer(
	ethClientFactory factory_interfaces.EthClientFactoryInterface,
	transactorFactory factory_interfaces.TransactorFactoryInterface,
) *Permit2ApprovalEnsurer {
	return &Permit2ApprovalEnsurer{
		ethClientFactory:  ethClientFactory,
		transactorFactory: transactorFactory,
	}
}

func (e *Permit2ApprovalEnsurer) EnsureApproval(
	ctx context.Context,
	network string,
	asset string,
	amount *big.Int,
) error {
	e.approvalMutex.Lock()
	defer e.approvalMutex.Unlock()

	evmChainId, err := helpers.ParseCaip2(network)
	if err != nil {
		return err
	}
	chainId := fmt.Sprintf("%d", evmChainId)

	client, err := e.ethClientFactory.For(chainId)
	if err != nil {
		return err
	}
	token, err := abi.NewErc20(common.HexToAddress(asset), client)
	if err != nil {
		return fmt.Errorf("failed to bind ERC-20 at %s: %w", asset, err)
	}

	transactor, err := e.transactorFactory.Create(ctx, evmChainId)
	if err != nil {
		return err
	}
	permit2 := common.HexToAddress(x402.Permit2ContractAddress)

	allowance, err := token.Allowance(&bind.CallOpts{Context: ctx}, transactor.From, permit2)
	if err != nil {
		return fmt.Errorf("failed to read permit2 allowance: %w", err)
	}
	if allowance != nil && allowance.Cmp(amount) >= 0 {
		return nil
	}

	// Tokens like USDT revert on a non-zero to non-zero allowance change.
	if allowance != nil && allowance.Sign() > 0 {
		reset, err := token.Approve(transactor, permit2, big.NewInt(0))
		if err != nil {
			return fmt.Errorf("failed to clear the existing permit2 allowance: %w", err)
		}
		if err := helper.WaitForSuccess(ctx, client, reset); err != nil {
			return err
		}
	}

	tx, err := token.Approve(transactor, permit2, maxUint256)
	if err != nil {
		return fmt.Errorf("failed to approve permit2: %w", err)
	}
	return helper.WaitForSuccess(ctx, client, tx)
}
