package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
)

// RegistrationReconcilerInterface persists confirmed registration facts into the
// agent + state rows. Shared by the check and self-register use cases.
type RegistrationReconcilerInterface interface {
	// ReconcileMint reads the freshly minted token's creator/owner from chain and
	// writes the tokenId + confirmed mint status into the agent and state rows.
	ReconcileMint(
		ctx context.Context,
		agent *models.Agent,
		state *models.AgentRegistrationState,
		collection common.Address,
		tokenId uint64,
		mintTxHash *string,
	) (*models.Agent, *models.AgentRegistrationState, error)
	// ReconcileEnsRegistry flips the ENS lifecycle to confirmed and persists it under the agent.
	ReconcileEnsRegistry(
		ctx context.Context,
		agentId uuid.UUID,
		state *models.AgentRegistrationState,
	) (*models.AgentRegistrationState, error)
}
