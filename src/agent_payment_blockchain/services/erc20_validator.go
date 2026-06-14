package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment_blockchain/abi"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment_blockchain/abi/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment_blockchain/helper"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

// Erc20Validator probes decimals/symbol/totalSupply with decode-checked reads
// and resolves the token's EIP-712 domain. EIP-3009 support is not statically
// detectable — the facilitator's verify is the gate at payment time.
type Erc20Validator struct {
	erc20Caller interfaces.Erc20CallerInterface
}

func NewErc20Validator(erc20Caller interfaces.Erc20CallerInterface) *Erc20Validator {
	return &Erc20Validator{erc20Caller: erc20Caller}
}

func (v *Erc20Validator) ValidateErc20(ctx context.Context, chainId int64, asset string) (enums.AssetTransferMethod, string, string, error) {
	token, err := abi.NewErc20Caller(common.HexToAddress(asset), v.erc20Caller)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to bind ERC-20 at %s: %w", asset, err)
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
				return "", "", "", fmt.Errorf("no contract deployed at %s on chain %d", asset, chainId)
			}
			return "", "", "", fmt.Errorf("no ERC-20 token answers %s at %s on chain %d", probe.name, asset, chainId)
		}
	}

	method, name, version := v.detectTransferMethod(token, chainId, asset, callOpts)
	return method, name, version, nil
}

// detectTransferMethod resolves the token's EIP-712 domain to decide the x402
// transfer method: a usable domain (EIP-5267 eip712Domain(), or name()+version()
// verified against DOMAIN_SEPARATOR()) means EIP-3009; otherwise the token goes
// through Permit2 (no per-token domain).
func (v *Erc20Validator) detectTransferMethod(token *abi.Erc20Caller, chainId int64, asset string, callOpts *bind.CallOpts) (enums.AssetTransferMethod, string, string) {
	if domain, err := token.Eip712Domain(callOpts); err == nil {
		return enums.AssetTransferMethodEip3009, domain.Name, domain.Version
	}

	name, nameErr := token.Name(callOpts)
	version, versionErr := token.Version(callOpts)
	if nameErr == nil && versionErr == nil {
		if separator, err := token.DOMAINSEPARATOR(callOpts); err == nil && helper.ComputeDomainSeparator(name, version, chainId, asset) == separator {
			return enums.AssetTransferMethodEip3009, name, version
		}
	}
	return enums.AssetTransferMethodPermit2, "", ""
}
