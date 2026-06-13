package use_cases

import (
	"context"

	identity_repositories "github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	runtime_common_settings "github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
)

type CreatePaymentRule struct {
	*AbstractPaymentUseCase
}

func NewCreatePaymentRule(
	agentRepository identity_repositories.AgentRepositoryInterface,
	paymentPolicyRepository repositories.PaymentPolicyRepositoryInterface,
	paymentRuleRepository repositories.PaymentRuleRepositoryInterface,
	paymentMethodStrategies payment_interfaces.PaymentMethodStrategiesInterface,
	blockchainSettings *runtime_common_settings.BlockchainSettings,
	token6022Settings *runtime_common_settings.Token6022Settings,
) *CreatePaymentRule {
	return &CreatePaymentRule{
		AbstractPaymentUseCase: NewAbstractPaymentUseCase(
			agentRepository,
			paymentPolicyRepository,
			paymentRuleRepository,
			paymentMethodStrategies,
			blockchainSettings,
			token6022Settings,
		),
	}
}

func (u *CreatePaymentRule) Execute(ctx context.Context, request *requests.SavePaymentRuleRequest) (*responses.PaymentRuleResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}
	strategy, ok := u.paymentMethodStrategies.For(enums.PaymentMethod(request.Method))
	if !ok {
		return nil, errors.NewValidationError("method", "unsupported payment method")
	}
	if err := strategy.Validate(ctx, request); err != nil {
		return nil, err
	}

	policy, _, err := u.ensurePaymentPolicy(ctx)
	if err != nil {
		return nil, err
	}

	rule, err := u.paymentRuleRepository.Create(ctx, toPaymentRule(policy, request))
	if err != nil {
		return nil, errors.NewInternalError("create_failed", err.Error())
	}
	// Persist the method config after the rule (it needs the rule id); roll the
	// rule back if that fails so we never leave an orphan without config.
	if err := strategy.SaveConfig(ctx, rule.Id, request.Config); err != nil {
		_ = u.paymentRuleRepository.DeleteById(ctx, rule.Id)
		return nil, errors.NewInternalError("create_failed", err.Error())
	}
	return responses.ToPaymentRuleResponse(rule, request.Config), nil
}

func toPaymentRule(policy *models.PaymentPolicy, request *requests.SavePaymentRuleRequest) *models.PaymentRule {
	return &models.PaymentRule{
		PolicyId: policy.Id,
		Name:     request.Name,
		Price:    request.Price,
		Method:   enums.PaymentMethod(request.Method),
	}
}
