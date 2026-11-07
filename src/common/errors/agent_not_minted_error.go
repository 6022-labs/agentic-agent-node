package errors

import stderrors "errors"

// ErrAgentNotMinted means the agent has no identity NFT yet, so it cannot hold sessions.
var ErrAgentNotMinted = stderrors.New("agent identity is not minted yet")
