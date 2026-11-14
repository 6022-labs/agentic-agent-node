package use_cases

import (
	"context"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type DeletePaymentSpendPolicy struct {
	paymentSpendPolicyRepository repositories.PaymentSpendPolicyRepositoryInterface
	agentResolver                identity_interfaces.AgentResolverInterface
}

func NewDeletePaymentSpendPolicy(
	paymentSpendPolicyRepository repositories.PaymentSpendPolicyRepositoryInterface,
	agentResolver identity_interfaces.AgentResolverInterface,
) *DeletePaymentSpendPolicy {
	return &DeletePaymentSpendPolicy{
		paymentSpendPolicyRepository: paymentSpendPolicyRepository,
		agentResolver:                agentResolver,
	}
}

func (u *DeletePaymentSpendPolicy) Execute(ctx context.Context) error {
	agent, err := u.agentResolver.Get(ctx)
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
