package services

import (
	"context"
	"errors"
	"fmt"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"sync"
)

// agentScopePrefix is the parent path every agent's secrets are namespaced under.
const agentScopePrefix = "agents"

// AgentScopedKeyBuilder prefixes secret keys with the persisted agent id, cached
// after the first lookup since it never changes.
type AgentScopedKeyBuilder struct {
	agentResolver identity_interfaces.AgentResolverInterface

	mu     sync.Mutex
	cached string
}

func NewAgentScopedKeyBuilder(agentResolver identity_interfaces.AgentResolverInterface) *AgentScopedKeyBuilder {
	return &AgentScopedKeyBuilder{agentResolver: agentResolver}
}

func (b *AgentScopedKeyBuilder) ScopedKey(ctx context.Context, key string) (string, error) {
	scope, err := b.scope(ctx)
	if err != nil {
		return "", err
	}
	return b.ScopedKeyFor(scope, key), nil
}

func (b *AgentScopedKeyBuilder) ScopedKeyFor(agentId, key string) string {
	return fmt.Sprintf("%s/%s/%s", agentScopePrefix, agentId, key)
}

func (b *AgentScopedKeyBuilder) scope(ctx context.Context) (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.cached != "" {
		return b.cached, nil
	}

	agent, err := b.agentResolver.Get(ctx)
	if err != nil {
		return "", err
	}
	if agent == nil {
		return "", errors.New("agent scope unavailable: no agent record exists yet")
	}

	b.cached = agent.Id.String()
	return b.cached, nil
}
