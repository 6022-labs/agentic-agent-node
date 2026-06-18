package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	"github.com/ethereum/go-ethereum/common"
)

// MintResult is the outcome of a self-mint: the persisted rows plus whether the
// mint is still pending a moderator's approval of a proposal.
type MintResult struct {
	Agent           *models.Agent
	State           *models.AgentRegistrationState
	ProposalPending bool
}

// MintPublisherInterface mints the agent's NFT with the node wallet (the self
// path) and reads its on-chain mint state. Paired with EnsPublisherInterface.
type MintPublisherInterface interface {
	// SelfMint mints with the node wallet, choosing a direct mint or a mint
	// proposal based on the collection's moderators. A direct mint reconciles the
	// rows; a proposal records the ProposalId and leaves the mint submitted.
	SelfMint(ctx context.Context, agent *models.Agent, state *models.AgentRegistrationState) (*MintResult, error)
	// ResolveMintedToken finds the agent's minted token on chain (the node
	// wallet's token in the draft collection, else a manager-wide scan). Returns
	// tokenId 0 when the agent is not minted yet.
	ResolveMintedToken(ctx context.Context, agent *models.Agent) (common.Address, uint64, error)
}
