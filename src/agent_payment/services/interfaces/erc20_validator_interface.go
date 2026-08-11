package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
)

// Erc20ValidatorInterface checks a rule's asset is a real ERC-20 and detects
// how x402 can move it.
type Erc20ValidatorInterface interface {
	// ValidateErc20 errors when no ERC-20 responds at asset on the given chain.
	// For a token with a usable EIP-712 domain it returns eip3009 + that domain's
	// name/version; otherwise permit2 (empty name/version). The caller guarantees
	// the chain is reachable.
	ValidateErc20(
		ctx context.Context,
		chainId string,
		asset string,
	) (method enums.AssetTransferMethod, eip712Name, eip712Version string, err error)
}
