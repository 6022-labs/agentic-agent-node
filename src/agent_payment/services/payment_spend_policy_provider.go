package services

import (
	"context"

	identity_repositories "github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

// seededSpendRuleName is the catch-all a fresh agent starts with.
const seededSpendRuleName = "Anything, unbounded"

// seededSpendMaxAmount is max-uint256: the seed bounds nothing until an owner narrows it.
const seededSpendMaxAmount = "115792089237316195423570985008687907853269984665640564039457584007913129639935"

const everyPattern = ".*"

// PaymentSpendPolicyProvider materializes the spend policy lazily; unlike receive it needs no minted identity.
type PaymentSpendPolicyProvider struct {
	agentRepository              identity_repositories.AgentRepositoryInterface
	paymentSpendPolicyRepository repositories.PaymentSpendPolicyRepositoryInterface
	paymentSpendRuleRepository   repositories.PaymentSpendRuleRepositoryInterface
	paymentSpendMethodStrategies payment_interfaces.PaymentSpendMethodStrategiesInterface
}

func NewPaymentSpendPolicyProvider(
	agentRepository identity_repositories.AgentRepositoryInterface,
	paymentSpendPolicyRepository repositories.PaymentSpendPolicyRepositoryInterface,
	paymentSpendRuleRepository repositories.PaymentSpendRuleRepositoryInterface,
	paymentSpendMethodStrategies payment_interfaces.PaymentSpendMethodStrategiesInterface,
) *PaymentSpendPolicyProvider {
	return &PaymentSpendPolicyProvider{
		agentRepository:              agentRepository,
		paymentSpendPolicyRepository: paymentSpendPolicyRepository,
		paymentSpendRuleRepository:   paymentSpendRuleRepository,
		paymentSpendMethodStrategies: paymentSpendMethodStrategies,
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

	if err := p.seedDefaultRule(ctx, policy); err != nil {
		return nil, err
	}
	return policy, nil
}

// seedDefaultRule gives a fresh agent an unbounded catch-all, matching how the receive side seeds free owner access.
func (p *PaymentSpendPolicyProvider) seedDefaultRule(ctx context.Context, policy *models.PaymentSpendPolicy) error {
	rule, err := p.paymentSpendRuleRepository.Create(ctx, &models.PaymentSpendRule{
		PolicyId:  policy.Id,
		Name:      seededSpendRuleName,
		MaxAmount: seededSpendMaxAmount,
		Method:    enums.PaymentMethodX402,
	})
	if err != nil {
		return errors.NewInternalError("spend_policy_failed", err.Error())
	}

	config := &models.PaymentSpendRuleX402Config{
		NetworkPatterns:             []string{everyPattern},
		AssetPatterns:               []string{everyPattern},
		PayToPatterns:               []string{everyPattern},
		AssetTransferMethodPatterns: []string{everyPattern},
	}
	raw, err := config.Marshal()
	if err != nil {
		return errors.NewInternalError("spend_policy_failed", err.Error())
	}

	strategy, ok := p.paymentSpendMethodStrategies.For(rule.Method)
	if !ok {
		return errors.NewInternalError("spend_policy_failed", "no strategy for the seeded rule's method")
	}
	if err := strategy.SaveConfig(ctx, rule.Id, raw); err != nil {
		return errors.NewInternalError("spend_policy_failed", err.Error())
	}
	return nil
}
