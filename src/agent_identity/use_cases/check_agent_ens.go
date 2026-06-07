package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/responses"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
)

// CheckAgentEns reconciles ENS provisioning from chain, regardless of who
// sent initAgentTexts. Idempotent: it reads the DB first and only hits the
// chain when the agent is minted but ENS is not yet confirmed.
type CheckAgentEns struct {
	ensPublisher                     identity_interfaces.EnsPublisherInterface
	agentRepository                  repositories.AgentRepositoryInterface
	agentRegistrationStateRepository repositories.AgentRegistrationStateRepositoryInterface
	registrationReconciler           identity_interfaces.RegistrationReconcilerInterface
}

func NewCheckAgentEns(
	ensPublisher identity_interfaces.EnsPublisherInterface,
	agentRepository repositories.AgentRepositoryInterface,
	agentRegistrationStateRepository repositories.AgentRegistrationStateRepositoryInterface,
	registrationReconciler identity_interfaces.RegistrationReconcilerInterface,
) *CheckAgentEns {
	return &CheckAgentEns{
		ensPublisher:                     ensPublisher,
		agentRepository:                  agentRepository,
		agentRegistrationStateRepository: agentRegistrationStateRepository,
		registrationReconciler:           registrationReconciler,
	}
}

func (u *CheckAgentEns) Execute(ctx context.Context) (*responses.RegistrationStatusResponse, error) {
	agent, err := u.agentRepository.Get(ctx)
	if err != nil {
		return nil, apperrors.NewInternalError("agent_lookup_failed", err.Error())
	}
	if agent == nil || agent.CollectionTokenId == nil {
		return nil, apperrors.NewConflictError("not_minted", "agent NFT must be minted before ENS registration")
	}

	state, err := u.agentRegistrationStateRepository.Get(ctx)
	if err != nil {
		return nil, apperrors.NewInternalError("state_lookup_failed", err.Error())
	}

	// Already confirmed in the DB: idempotent, skip the chain read entirely.
	if state == nil || state.EnsRegistryStatus != enums.TransactionStatusConfirmed {
		provisioned, err := u.ensPublisher.IsProvisioned(ctx)
		if err != nil {
			return nil, apperrors.NewUnavailableError("ens_read_failed", err.Error())
		}
		if provisioned {
			state, err = u.registrationReconciler.ReconcileEnsRegistry(ctx, state)
			if err != nil {
				return nil, apperrors.NewInternalError("ens_reconcile_failed", err.Error())
			}
		}
	}

	resp := &responses.RegistrationStatusResponse{}
	resp.FromModel(agent, state)
	return resp, nil
}
