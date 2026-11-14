package use_cases

import (
	"context"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type DeletePaymentReceivePolicy struct {
	paymentReceivePolicyRepository repositories.PaymentReceivePolicyRepositoryInterface
	agentResolver                  identity_interfaces.AgentResolverInterface
}

func NewDeletePaymentReceivePolicy(
	paymentReceivePolicyRepository repositories.PaymentReceivePolicyRepositoryInterface,
	agentResolver identity_interfaces.AgentResolverInterface,
) *DeletePaymentReceivePolicy {
	return &DeletePaymentReceivePolicy{
		paymentReceivePolicyRepository: paymentReceivePolicyRepository,
		agentResolver:                  agentResolver,
	}
}

func (u *DeletePaymentReceivePolicy) Execute(ctx context.Context) error {
	agent, err := u.agentResolver.Get(ctx)
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
