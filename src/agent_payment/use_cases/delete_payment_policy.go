package use_cases

import (
	"context"

	identity_repositories "github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type DeletePaymentPolicy struct {
	paymentPolicyRepository repositories.PaymentPolicyRepositoryInterface
	agentRepository         identity_repositories.AgentRepositoryInterface
}

func NewDeletePaymentPolicy(
	paymentPolicyRepository repositories.PaymentPolicyRepositoryInterface,
	agentRepository identity_repositories.AgentRepositoryInterface,
) *DeletePaymentPolicy {
	return &DeletePaymentPolicy{paymentPolicyRepository: paymentPolicyRepository, agentRepository: agentRepository}
}

func (u *DeletePaymentPolicy) Execute(ctx context.Context) error {
	agent, err := u.agentRepository.Get(ctx)
	if err != nil {
		return errors.NewInternalError("delete_failed", err.Error())
	}
	if agent == nil {
		return errors.NewInternalError("delete_failed", "agent identity is not initialized yet")
	}
	if err := u.paymentPolicyRepository.DeleteByAgentId(ctx, agent.Id); err != nil {
		return errors.NewInternalError("delete_failed", err.Error())
	}
	return nil
}
