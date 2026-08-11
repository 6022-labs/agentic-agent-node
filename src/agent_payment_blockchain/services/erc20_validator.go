package services

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment_blockchain/abi"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment_blockchain/factory/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment_blockchain/helper"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

// Erc20Validator probes decimals/symbol/totalSupply with decode-checked reads
// and resolves the token's EIP-712 domain. EIP-3009 support is not statically
// detectable — the facilitator's verify is the gate at payment time. It resolves
// the client per chain so it can validate assets on any reachable chain.
type Erc20Validator struct {
	ethClientFactory interfaces.EthClientFactoryInterface
}

func NewErc20Validator(ethClientFactory interfaces.EthClientFactoryInterface) *Erc20Validator {
	return &Erc20Validator{ethClientFactory: ethClientFactory}
}

func (v *Erc20Validator) ValidateErc20(
	ctx context.Context,
	chainId string,
	asset string,
) (enums.AssetTransferMethod, string, string, error) {
	// Numeric id for the EIP-712 domain separator.
	evmChainId, err := strconv.ParseUint(chainId, 10, 64)
	if err != nil {
		return "", "", "", err
	}
	client, err := v.ethClientFactory.For(chainId)
	if err != nil {
		return "", "", "", err
	}
	token, err := abi.NewErc20Caller(common.HexToAddress(asset), client)
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
				return "", "", "", fmt.Errorf("no contract deployed at %s on chain %s", asset, chainId)
			}
			return "", "", "", fmt.Errorf("no ERC-20 token answers %s at %s on chain %s", probe.name, asset, chainId)
		}
	}

	method, name, version := v.detectTransferMethod(token, int64(evmChainId), asset, callOpts)
	return method, name, version, nil
}

// detectTransferMethod resolves the token's EIP-712 domain to decide the x402
// transfer method: a usable domain (EIP-3009 or Permit2).
func (v *Erc20Validator) detectTransferMethod(
	token *abi.Erc20Caller,
	chainId int64,
	asset string,
	callOpts *bind.CallOpts,
) (enums.AssetTransferMethod, string, string) {
	if domain, err := token.Eip712Domain(callOpts); err == nil {
		return enums.AssetTransferMethodEip3009, domain.Name, domain.Version
	}

	name, nameErr := token.Name(callOpts)
	version, versionErr := token.Version(callOpts)
	if nameErr == nil && versionErr == nil {
		if separator, err := token.DOMAINSEPARATOR(callOpts); err == nil &&
			helper.ComputeDomainSeparator(name, version, chainId, asset) == separator {
			return enums.AssetTransferMethodEip3009, name, version
		}
	}
	return enums.AssetTransferMethodPermit2, "", ""
}
