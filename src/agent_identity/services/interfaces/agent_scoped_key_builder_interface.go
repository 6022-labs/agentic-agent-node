package interfaces

import "context"

// AgentScopedKeyBuilderInterface namespaces a secret key to the current agent so
// multiple agent-nodes can share one Vault without colliding.
type AgentScopedKeyBuilderInterface interface {
	// ScopedKey prefixes key with the agent's namespace; errors if no agent exists yet.
	ScopedKey(ctx context.Context, key string) (string, error)
}
