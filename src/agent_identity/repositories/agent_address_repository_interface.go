package repositories

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	"github.com/google/uuid"
)

// AgentAddressRepositoryInterface persists the agent's wallet addresses, one row
// per chain family (agent_id, type). The private keys live in Vault.
type AgentAddressRepositoryInterface interface {
	// Upsert replaces the address row for (agentId, type).
	Upsert(ctx context.Context, address *models.AgentAddress) error
	// GetByAgentIdAndType returns the address, or (nil, nil) if none exists.
	GetByAgentIdAndType(
		ctx context.Context,
		agentId uuid.UUID,
		addressType enums.AddressType,
	) (*models.AgentAddress, error)
	// GetAllByAgentId returns the agent's addresses keyed by type.
	GetAllByAgentId(ctx context.Context, agentId uuid.UUID) (map[enums.AddressType]string, error)
}
