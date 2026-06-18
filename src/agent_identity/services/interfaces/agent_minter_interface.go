package interfaces

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
)

// MintKeyValue is one image/attribute entry for a mint call.
type MintKeyValue struct {
	Key   string
	Value string
}

// FromMap builds a MintKeyValue slice from a string map.
func (MintKeyValue) FromMap(m map[string]string) []MintKeyValue {
	out := make([]MintKeyValue, 0, len(m))
	for k, v := range m {
		out = append(out, MintKeyValue{Key: k, Value: v})
	}
	return out
}

// MintParams is the mint() / createMintProposal() payload; Wallets includes the
// node wallet.
type MintParams struct {
	Collection common.Address
	To         common.Address
	Role       string
	Name       string
	Wallets    []common.Address
	Images     []MintKeyValue
	Attributes []MintKeyValue
	CloneOf    *uint64
}

// AgentMinterInterface writes mint transactions with the node's own wallet, on
// the given chain.
type AgentMinterInterface interface {
	// Mint calls mint() directly (no moderators, or node wallet is one). Returns the tx hash.
	Mint(ctx context.Context, chainId string, params MintParams) (txHash string, err error)
	// CreateMintProposal calls createMintProposal() on moderated collections. Returns tx hash + proposalId.
	CreateMintProposal(ctx context.Context, chainId string, params MintParams) (txHash string, proposalId uint64, err error)
}
