package errors

import stderrors "errors"

// ErrConversationUnavailable means the requested conversation key belongs to another caller.
var ErrConversationUnavailable = stderrors.New("conversation unavailable for caller")
