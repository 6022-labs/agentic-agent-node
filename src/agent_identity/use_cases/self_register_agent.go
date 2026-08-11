package use_cases

import (
	"context"

	metrics_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/metrics/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/responses"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
)

// SelfRegisterAgent mints + ENS-registers via the node wallet (after the owner
// funds it), reusing the stored draft. The mint may land directly or be left
// pending a moderator's approval of a proposal.
type SelfRegisterAgent struct {
	agentRepository                  repositories.AgentRepositoryInterface
	agentRegistrationStateRepository repositories.AgentRegistrationStateRepositoryInterface
	mintPublisher                    identity_interfaces.MintPublisherInterface
	ensPublisher                     identity_interfaces.EnsPublisherInterface
	registrationReconciler           identity_interfaces.RegistrationReconcilerInterface
	registrationMetrics              metrics_interfaces.RegistrationMetricsInterface
}

func NewSelfRegisterAgent(
	agentRepository repositories.AgentRepositoryInterface,
	agentRegistrationStateRepository repositories.AgentRegistrationStateRepositoryInterface,
	mintPublisher identity_interfaces.MintPublisherInterface,
	ensPublisher identity_interfaces.EnsPublisherInterface,
	registrationReconciler identity_interfaces.RegistrationReconcilerInterface,
	registrationMetrics metrics_interfaces.RegistrationMetricsInterface,
) *SelfRegisterAgent {
	return &SelfRegisterAgent{
		agentRepository:                  agentRepository,
		agentRegistrationStateRepository: agentRegistrationStateRepository,
		mintPublisher:                    mintPublisher,
		ensPublisher:                     ensPublisher,
		registrationReconciler:           registrationReconciler,
		registrationMetrics:              registrationMetrics,
	}
}

func (u *SelfRegisterAgent) Execute(ctx context.Context) (*responses.RegistrationStatusResponse, error) {
	succeeded, pending := false, false
	defer func() {
		switch {
		case pending:
			u.registrationMetrics.RecordRegistrationPending(ctx)
		case succeeded:
			u.registrationMetrics.RecordRegistrationSucceeded(ctx)
		default:
			u.registrationMetrics.RecordRegistrationFailed(ctx)
		}
	}()

	agent, err := u.agentRepository.Get(ctx)
	if err != nil {
		return nil, apperrors.NewInternalError("agent_lookup_failed", err.Error())
	}
	if agent == nil {
		return nil, apperrors.NewNotFoundError("draft_not_found", "no agent registration has been saved yet")
	}
	if agent.Owner == nil {
		return nil, apperrors.NewConflictError("owner_missing", "agent draft has no owner")
	}

	state, err := u.agentRegistrationStateRepository.GetByAgentId(ctx, agent.Id)
	if err != nil {
		return nil, apperrors.NewInternalError("state_lookup_failed", err.Error())
	}

	agent, state, pending, err = u.ensureMinted(ctx, agent, state)
	if err != nil {
		return nil, err
	}
	if pending {
		resp := &responses.RegistrationStatusResponse{}
		resp.FromModel(agent, state)
		return resp, nil
	}

	// ENS-register with the node wallet unless already provisioned.
	if state == nil || state.EnsRegistryStatus != enums.TransactionStatusConfirmed {
		if err := u.ensPublisher.PublishRuntimeRecords(ctx); err != nil {
			return nil, apperrors.NewInternalError("self_ens_failed", err.Error())
		}
		state, err = u.registrationReconciler.ReconcileEnsRegistry(ctx, agent.Id, state)
		if err != nil {
			return nil, apperrors.NewInternalError("ens_reconcile_failed", err.Error())
		}
	}

	succeeded = true
	resp := &responses.RegistrationStatusResponse{}
	resp.FromModel(agent, state)
	return resp, nil
}

// ensureMinted reconciles an already-minted token so /self is resumable; the bool reports a pending proposal.
func (u *SelfRegisterAgent) ensureMinted(
	ctx context.Context,
	agent *models.Agent,
	state *models.AgentRegistrationState,
) (*models.Agent, *models.AgentRegistrationState, bool, error) {
	if agent.CollectionTokenId != nil {
		return agent, state, false, nil
	}

	collection, tokenId, err := u.mintPublisher.ResolveMintedToken(ctx, agent)
	if err != nil {
		return nil, nil, false, apperrors.NewUnavailableError("chain_read_failed", err.Error())
	}

	if tokenId != 0 {
		agent, state, err = u.registrationReconciler.ReconcileMint(ctx, agent, state, collection, tokenId, nil)
		if err != nil {
			return nil, nil, false, apperrors.NewInternalError("mint_reconcile_failed", err.Error())
		}
		return agent, state, false, nil
	}

	result, err := u.mintPublisher.SelfMint(ctx, agent, state)
	if err != nil {
		return nil, nil, false, apperrors.NewInternalError("self_mint_failed", err.Error())
	}
	return result.Agent, result.State, result.ProposalPending, nil
}
