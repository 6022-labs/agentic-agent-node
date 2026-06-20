package use_cases

import (
	"context"

	metrics_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/metrics/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

// CheckAgentMint reconciles the agent's mint status from chain, regardless of who
// sent the mint (browser wallet or the node itself). Idempotent.
type CheckAgentMint struct {
	agentRepository                  repositories.AgentRepositoryInterface
	agentRegistrationStateRepository repositories.AgentRegistrationStateRepositoryInterface
	mintPublisher                    identity_interfaces.MintPublisherInterface
	registrationReconciler           identity_interfaces.RegistrationReconcilerInterface
	mintMetrics                      metrics_interfaces.MintMetricsInterface
}

func NewCheckAgentMint(
	agentRepository repositories.AgentRepositoryInterface,
	agentRegistrationStateRepository repositories.AgentRegistrationStateRepositoryInterface,
	mintPublisher identity_interfaces.MintPublisherInterface,
	registrationReconciler identity_interfaces.RegistrationReconcilerInterface,
	mintMetrics metrics_interfaces.MintMetricsInterface,
) *CheckAgentMint {
	return &CheckAgentMint{
		agentRepository:                  agentRepository,
		agentRegistrationStateRepository: agentRegistrationStateRepository,
		mintPublisher:                    mintPublisher,
		registrationReconciler:           registrationReconciler,
		mintMetrics:                      mintMetrics,
	}
}

func (u *CheckAgentMint) Execute(ctx context.Context) (*responses.RegistrationStatusResponse, error) {
	agent, err := u.agentRepository.Get(ctx)
	if err != nil {
		return nil, errors.NewInternalError("agent_lookup_failed", err.Error())
	}
	if agent == nil {
		return nil, errors.NewNotFoundError("draft_not_found", "no agent registration has been saved yet")
	}

	state, err := u.agentRegistrationStateRepository.GetByAgentId(ctx, agent.Id)
	if err != nil {
		return nil, errors.NewInternalError("state_lookup_failed", err.Error())
	}

	// Already minted in the DB: idempotent, skip the chain read entirely.
	if agent.CollectionTokenId != nil {
		resp := &responses.RegistrationStatusResponse{}
		resp.FromModel(agent, state)
		return resp, nil
	}

	collection, tokenId, err := u.mintPublisher.ResolveMintedToken(ctx, agent)
	if err != nil {
		return nil, errors.NewUnavailableError("chain_read_failed", err.Error())
	}

	// Not minted yet: return the current stored status untouched.
	if tokenId == 0 {
		resp := &responses.RegistrationStatusResponse{}
		resp.FromModel(agent, state)
		return resp, nil
	}

	savedAgent, savedState, err := u.registrationReconciler.ReconcileMint(ctx, agent, state, collection, tokenId, nil)
	if err != nil {
		u.mintMetrics.RecordMintFailed(ctx)
		return nil, errors.NewInternalError("mint_reconcile_failed", err.Error())
	}
	u.mintMetrics.RecordMintSucceeded(ctx)
	resp := &responses.RegistrationStatusResponse{}
	resp.FromModel(savedAgent, savedState)
	return resp, nil
}
