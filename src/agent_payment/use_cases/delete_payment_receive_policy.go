package use_cases

import (
	"context"

	identity_repositories "github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type DeletePaymentReceivePolicy struct {
	paymentReceivePolicyRepository repositories.PaymentReceivePolicyRepositoryInterface
	agentRepository                identity_repositories.AgentRepositoryInterface
}

func NewDeletePaymentReceivePolicy(
	paymentReceivePolicyRepository repositories.PaymentReceivePolicyRepositoryInterface,
	agentRepository identity_repositories.AgentRepositoryInterface,
) *DeletePaymentReceivePolicy {
	return &DeletePaymentReceivePolicy{
		paymentReceivePolicyRepository: paymentReceivePolicyRepository,
		agentRepository:                agentRepository,
	}
}

func (u *DeletePaymentReceivePolicy) Execute(ctx context.Context) error {
	agent, err := u.agentRepository.Get(ctx)
	if err != nil {
		return errors.NewInternalError("delete_failed", err)
	}
	if agent == nil {
		return errors.NewInternalError("delete_failed", errors.ErrAgentNotInitialized)
	}
	if err := u.paymentReceivePolicyRepository.DeleteByAgentId(ctx, agent.Id); err != nil {
		return errors.NewInternalError("delete_failed", err)
	}
	return nil
}
