package interfaces

import "context"

// Erc20ValidatorInterface checks that a rule's asset is a real ERC-20.
type Erc20ValidatorInterface interface {
	// ValidateErc20 errors when no ERC-20 responds at asset on the given
	// chain; chains other than the node's are skipped (no RPC for them).
	ValidateErc20(ctx context.Context, chainId int64, asset string) error
}
