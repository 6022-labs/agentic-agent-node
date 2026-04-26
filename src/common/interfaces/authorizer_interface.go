package interfaces

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
)

// AuthorizerInterface decides whether a wallet address is allowed to act on
// behalf of this agent (e.g. configure secrets, view privileged state).
//
// v1 is satisfied by NftOwnerAuthorizer, which checks the address against the
// current NFT owner. Future implementations (AgentSelfAuthorizer, composite
// AnyOf) compose around the same shape.
type AuthorizerInterface interface {
	IsAuthorized(ctx context.Context, address common.Address) (bool, error)
}
