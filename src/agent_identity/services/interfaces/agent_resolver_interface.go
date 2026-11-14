package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
)

// AgentResolverInterface is the read path to the agent row for every domain.
type AgentResolverInterface interface {
	// Get returns the persisted agent, or (nil, nil) if none, without waiting on chain.
	// A stale minted identity is resynced in the background.
	Get(ctx context.Context) (*models.Agent, error)
}
