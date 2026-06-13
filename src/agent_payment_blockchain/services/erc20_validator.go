package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment_blockchain/abi"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment_blockchain/abi/interfaces"
	runtime_settings "github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

// Erc20Validator probes decimals/symbol/totalSupply with decode-checked
// reads. EIP-3009 support is not statically detectable — the facilitator's
// verify is the gate at payment time.
type Erc20Validator struct {
	erc20Caller        interfaces.Erc20CallerInterface
	blockchainSettings *runtime_settings.BlockchainSettings
}

func NewErc20Validator(
	erc20Caller interfaces.Erc20CallerInterface,
	blockchainSettings *runtime_settings.BlockchainSettings,
) *Erc20Validator {
	return &Erc20Validator{
		erc20Caller:        erc20Caller,
		blockchainSettings: blockchainSettings,
	}
}

func (v *Erc20Validator) ValidateErc20(ctx context.Context, chainId int64, asset string) error {
	if chainId <= 0 || uint64(chainId) != v.blockchainSettings.Id {
		return nil
	}

	token, err := abi.NewErc20Caller(common.HexToAddress(asset), v.erc20Caller)
	if err != nil {
		return fmt.Errorf("failed to bind ERC-20 at %s: %w", asset, err)
	}

	callOpts := &bind.CallOpts{Context: ctx}
	probes := []struct {
		name string
		call func() error
	}{
		{"decimals()", func() error { _, err := token.Decimals(callOpts); return err }},
		{"symbol()", func() error { _, err := token.Symbol(callOpts); return err }},
		{"totalSupply()", func() error { _, err := token.TotalSupply(callOpts); return err }},
	}
	for _, probe := range probes {
		if err := probe.call(); err != nil {
			if errors.Is(err, bind.ErrNoCode) {
				return fmt.Errorf("no contract deployed at %s on chain %d", asset, chainId)
			}
			return fmt.Errorf("no ERC-20 token answers %s at %s on chain %d", probe.name, asset, chainId)
		}
	}
	return nil
}
