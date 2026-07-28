// Package fixtures deploys the tokens the ERC-20 validation tests probe; see README.md.
package fixtures

import (
	"context"
	"embed"
	"fmt"

	"github.com/6022-labs/agentic-agent-node/tests/common_tests/container_builders"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/contracts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

//go:embed artifacts/*.json
var artifactFiles embed.FS

// Erc20WithoutDomain has no EIP-712 domain, so the validator falls back to permit2.
const Erc20WithoutDomain = "TestERC20"

// Erc20WithDomain exposes an ERC-5267 eip712Domain, which the validator reads as EIP-3009.
const Erc20WithDomain = "TestERC20Permit"

// DeployErc20 deploys one of the token doubles.
func DeployErc20(
	ctx context.Context,
	client *ethclient.Client,
	anvilAccountSettings container_builders.AnvilAccountSettings,
	contractName string,
	tokenName string,
	tokenSymbol string,
) (common.Address, error) {
	deployer, err := contracts.NewDeployer(ctx, client, anvilAccountSettings)
	if err != nil {
		return common.Address{}, err
	}
	contract, err := load(contractName)
	if err != nil {
		return common.Address{}, err
	}

	address, _, err := deployer.Deploy(contractName, contract, tokenName, tokenSymbol)
	return address, err
}

func load(name string) (*contracts.CompiledContract, error) {
	raw, err := artifactFiles.ReadFile("artifacts/" + name + ".json")
	if err != nil {
		return nil, fmt.Errorf("no committed artifact for %s (see README.md): %w", name, err)
	}
	return contracts.Parse(raw)
}
