package use_cases

import (
	"context"

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
}

func NewSelfRegisterAgent(
	agentRepository repositories.AgentRepositoryInterface,
	agentRegistrationStateRepository repositories.AgentRegistrationStateRepositoryInterface,
	mintPublisher identity_interfaces.MintPublisherInterface,
	ensPublisher identity_interfaces.EnsPublisherInterface,
	registrationReconciler identity_interfaces.RegistrationReconcilerInterface,
) *SelfRegisterAgent {
	return &SelfRegisterAgent{
		agentRepository:                  agentRepository,
		agentRegistrationStateRepository: agentRegistrationStateRepository,
		mintPublisher:                    mintPublisher,
		ensPublisher:                     ensPublisher,
		registrationReconciler:           registrationReconciler,
	}
}

func (u *SelfRegisterAgent) Execute(ctx context.Context) (*responses.RegistrationStatusResponse, error) {
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

	state, err := u.agentRegistrationStateRepository.Get(ctx)
	if err != nil {
		return nil, apperrors.NewInternalError("state_lookup_failed", err.Error())
	}

	// Mint with the node wallet unless already minted.
	if agent.CollectionTokenId == nil {
		result, mintErr := u.mintPublisher.SelfMint(ctx, agent, state)
		if mintErr != nil {
			return nil, apperrors.NewInternalError("self_mint_failed", mintErr.Error())
		}
		if result.ProposalPending {
			resp := &responses.RegistrationStatusResponse{}
			resp.FromModel(result.Agent, result.State)
			return resp, nil
		}
		agent = result.Agent
		state = result.State
	}

	// ENS-register with the node wallet unless already provisioned.
	if state == nil || state.EnsRegistryStatus != enums.TransactionStatusConfirmed {
		if err := u.ensPublisher.PublishRuntimeRecords(ctx); err != nil {
			return nil, apperrors.NewInternalError("self_ens_failed", err.Error())
		}
		state, err = u.registrationReconciler.ReconcileEnsRegistry(ctx, state)
		if err != nil {
			return nil, apperrors.NewInternalError("ens_reconcile_failed", err.Error())
		}
	}

	resp := &responses.RegistrationStatusResponse{}
	resp.FromModel(agent, state)
	return resp, nil
}
