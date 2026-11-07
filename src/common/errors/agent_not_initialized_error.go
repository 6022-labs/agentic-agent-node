package errors

import stderrors "errors"

// ErrAgentNotInitialized means the agent identity has not been bootstrapped yet.
var ErrAgentNotInitialized = stderrors.New("agent identity is not initialized yet")
