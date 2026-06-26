package observability

import (
	"context"
	"strings"

	"go.opentelemetry.io/otel/attribute"
)

// AgentIdentity is the agent identity attached to telemetry signals; new fields flow to every signal via attributes.
type AgentIdentity struct {
	EnsDomain string
}

type agentIdentityContextKey struct{}

// ContextWithAgentIdentity carries the agent identity so the sinks tag the right agent (one or many).
func ContextWithAgentIdentity(ctx context.Context, identity AgentIdentity) context.Context {
	return context.WithValue(ctx, agentIdentityContextKey{}, identity)
}

func AgentIdentityFromContext(ctx context.Context) (AgentIdentity, bool) {
	identity, ok := ctx.Value(agentIdentityContextKey{}).(AgentIdentity)
	return identity, ok
}

// attributes renders the identity as OTel attributes; empty fields are skipped.
func (i AgentIdentity) attributes() []attribute.KeyValue {
	var attrs []attribute.KeyValue
	if strings.TrimSpace(i.EnsDomain) != "" {
		attrs = append(attrs, attribute.String("agent.ens.domain", i.EnsDomain))
	}
	return attrs
}
