package interfaces

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
)

// AuthorizerInterface decides whether a wallet address may act on behalf of
// this agent. v1: NftOwnerAuthorizer checks the address against the NFT owner.
type AuthorizerInterface interface {
	IsAuthorized(ctx context.Context, address common.Address) (bool, error)
}
