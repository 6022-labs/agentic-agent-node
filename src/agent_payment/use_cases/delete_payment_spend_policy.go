package use_cases

import (
	"context"

	identity_repositories "github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type DeletePaymentSpendPolicy struct {
	paymentSpendPolicyRepository repositories.PaymentSpendPolicyRepositoryInterface
	agentRepository              identity_repositories.AgentRepositoryInterface
}

func NewDeletePaymentSpendPolicy(
	paymentSpendPolicyRepository repositories.PaymentSpendPolicyRepositoryInterface,
	agentRepository identity_repositories.AgentRepositoryInterface,
) *DeletePaymentSpendPolicy {
	return &DeletePaymentSpendPolicy{
		paymentSpendPolicyRepository: paymentSpendPolicyRepository,
		agentRepository:              agentRepository,
	}
}

func (u *DeletePaymentSpendPolicy) Execute(ctx context.Context) error {
	agent, err := u.agentRepository.Get(ctx)
	if err != nil {
		return errors.NewInternalError("delete_failed", err)
	}
	// No agent yet means no policy either, so there is nothing to delete.
	if agent == nil {
		return nil
	}
	if err := u.paymentSpendPolicyRepository.DeleteByAgentId(ctx, agent.Id); err != nil {
		return errors.NewInternalError("delete_failed", err)
	}
	return nil
}
