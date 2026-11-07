package errors

import stderrors "errors"

// ErrIdentityNotFound means no NFT is minted to the agent's wallet yet.
var ErrIdentityNotFound = stderrors.New("agent identity not found on chain")
