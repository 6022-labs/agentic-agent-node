package use_cases

import (
	"context"

	identity_repositories "github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	runtime_common_settings "github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
)

type ListPaymentRules struct {
	*AbstractPaymentUseCase
}

func NewListPaymentRules(
	agentRepository identity_repositories.AgentRepositoryInterface,
	paymentPolicyRepository repositories.PaymentPolicyRepositoryInterface,
	paymentRuleRepository repositories.PaymentRuleRepositoryInterface,
	blockchainSettings *runtime_common_settings.BlockchainSettings,
	token6022Settings *runtime_common_settings.Token6022Settings,
	paymentFacilitatorsSettings *runtime_common_settings.PaymentFacilitatorsSettings,
) *ListPaymentRules {
	return &ListPaymentRules{
		AbstractPaymentUseCase: NewAbstractPaymentUseCase(
			agentRepository,
			paymentPolicyRepository,
			paymentRuleRepository,
			blockchainSettings,
			token6022Settings,
			paymentFacilitatorsSettings,
		),
	}
}

func (u *ListPaymentRules) Execute(ctx context.Context) (*responses.ListPaymentRulesResponse, error) {
	if _, _, err := u.ensurePaymentPolicy(ctx); err != nil {
		return nil, err
	}
	rules, err := u.paymentRuleRepository.List(ctx)
	if err != nil {
		return nil, errors.NewInternalError("list_failed", err.Error())
	}
	return &responses.ListPaymentRulesResponse{Rules: responses.ToPaymentRuleResponses(rules)}, nil
}
