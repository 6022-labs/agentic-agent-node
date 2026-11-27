package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/common/x402"
)

// A2AAgentCard is what the peer advertises about itself.
type A2AAgentCard struct {
	Name        string
	EndpointUrl string
}

// A2AReply is one answered turn from a peer agent.
type A2AReply struct {
	Text      string
	ContextId string
}

// A2ARequesterInterface talks A2A to another agent over HTTP.
type A2ARequesterInterface interface {
	// FetchAgentCard reads the peer's A2A card; any failure is an error the caller may fall back from.
	FetchAgentCard(ctx context.Context, baseUrl string) (*A2AAgentCard, error)
	// SendMessage runs one SendMessage turn. A priced endpoint answers with a quote and a nil reply;
	// paying is the caller's decision.
	SendMessage(
		ctx context.Context,
		endpointUrl string,
		message string,
		contextId string,
		paymentSignature string,
	) (reply *A2AReply, paymentRequired *x402.PaymentRequired, err error)
}
