package interfaces

import "errors"

// ErrIdentityNotFound means no NFT is minted to the agent's wallet yet.
var ErrIdentityNotFound = errors.New("agent identity not found on chain")
