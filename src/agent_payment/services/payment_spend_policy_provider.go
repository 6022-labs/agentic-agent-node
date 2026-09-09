package services

import (
	"context"

	identity_repositories "github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

// PaymentSpendPolicyProvider materializes the policy head only: an agent spends nothing until an owner writes a rule.
type PaymentSpendPolicyProvider struct {
	agentRepository              identity_repositories.AgentRepositoryInterface
	paymentSpendPolicyRepository repositories.PaymentSpendPolicyRepositoryInterface
}

func NewPaymentSpendPolicyProvider(
	agentRepository identity_repositories.AgentRepositoryInterface,
	paymentSpendPolicyRepository repositories.PaymentSpendPolicyRepositoryInterface,
) *PaymentSpendPolicyProvider {
	return &PaymentSpendPolicyProvider{
		agentRepository:              agentRepository,
		paymentSpendPolicyRepository: paymentSpendPolicyRepository,
	}
}

func (p *PaymentSpendPolicyProvider) Ensure(ctx context.Context) (*models.PaymentSpendPolicy, error) {
	agent, err := p.agentRepository.Get(ctx)
	if err != nil {
		return nil, errors.NewInternalError("spend_policy_failed", err.Error())
	}
	if agent == nil {
		return nil, errors.NewInternalError("spend_policy_failed", "agent identity is not initialized yet")
	}

	policy, err := p.paymentSpendPolicyRepository.GetByAgentId(ctx, agent.Id)
	if err != nil {
		return nil, errors.NewInternalError("spend_policy_failed", err.Error())
	}
	if policy != nil {
		return policy, nil
	}

	policy, err = p.paymentSpendPolicyRepository.Create(ctx, &models.PaymentSpendPolicy{AgentId: agent.Id})
	if err != nil {
		// A concurrent first request may have won the insert race; re-read before failing.
		existing, getErr := p.paymentSpendPolicyRepository.GetByAgentId(ctx, agent.Id)
		if getErr == nil && existing != nil {
			return existing, nil
		}
		return nil, errors.NewInternalError("spend_policy_failed", err.Error())
	}

	return policy, nil
}
