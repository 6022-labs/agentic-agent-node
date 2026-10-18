package helpers

import (
	"context"
	"net/http"

	common_helpers "github.com/6022-labs/agentic-agent-node/src/common/helpers"
)

// ForwardCallerId is the only place the gateway hands the authenticated caller to the runtime.
func ForwardCallerId(ctx context.Context, req *http.Request) error {
	if callerId, ok := common_helpers.CallerIdFromContext(ctx); ok {
		req.Header.Set(common_helpers.CallerIdHeader, callerId)
	}
	return nil
}
