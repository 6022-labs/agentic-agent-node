package services

import (
	"context"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

// PaymentSpendPolicyProvider materializes the policy head only: an agent spends nothing until an owner writes a rule.
type PaymentSpendPolicyProvider struct {
	agentResolver                identity_interfaces.AgentResolverInterface
	paymentSpendPolicyRepository repositories.PaymentSpendPolicyRepositoryInterface
}

func NewPaymentSpendPolicyProvider(
	agentResolver identity_interfaces.AgentResolverInterface,
	paymentSpendPolicyRepository repositories.PaymentSpendPolicyRepositoryInterface,
) *PaymentSpendPolicyProvider {
	return &PaymentSpendPolicyProvider{
		agentResolver:                agentResolver,
		paymentSpendPolicyRepository: paymentSpendPolicyRepository,
	}
}

func (p *PaymentSpendPolicyProvider) Ensure(ctx context.Context) (*models.PaymentSpendPolicy, error) {
	agent, err := p.agentResolver.Get(ctx)
	if err != nil {
		return nil, errors.NewInternalError("spend_policy_failed", err)
	}
	if agent == nil {
		return nil, errors.NewInternalError("spend_policy_failed", errors.ErrAgentNotInitialized)
	}
	// Before mint anyone can sign in to the console, so only a minted owner may configure spending.
	if agent.CollectionTokenId == nil {
		return nil, errors.NewConflictError("agent_not_minted", "agent identity is not minted yet")
	}

	policy, err := p.paymentSpendPolicyRepository.GetByAgentId(ctx, agent.Id)
	if err != nil {
		return nil, errors.NewInternalError("spend_policy_failed", err)
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
		return nil, errors.NewInternalError("spend_policy_failed", err)
	}

	return policy, nil
}
