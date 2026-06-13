package use_cases

import (
	"context"

	identity_repositories "github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
	runtime_common_settings "github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
)

type GetPaymentPolicy struct {
	*AbstractPaymentUseCase
}

func NewGetPaymentPolicy(
	agentRepository identity_repositories.AgentRepositoryInterface,
	paymentPolicyRepository repositories.PaymentPolicyRepositoryInterface,
	paymentRuleRepository repositories.PaymentRuleRepositoryInterface,
	blockchainSettings *runtime_common_settings.BlockchainSettings,
	token6022Settings *runtime_common_settings.Token6022Settings,
	paymentFacilitatorsSettings *runtime_common_settings.PaymentFacilitatorsSettings,
) *GetPaymentPolicy {
	return &GetPaymentPolicy{
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

func (u *GetPaymentPolicy) Execute(ctx context.Context) (*responses.PaymentPolicyResponse, error) {
	policy, agent, err := u.ensurePaymentPolicy(ctx)
	if err != nil {
		return nil, err
	}
	return responses.ToPaymentPolicyResponse(policy, agent.WalletAddress), nil
}
