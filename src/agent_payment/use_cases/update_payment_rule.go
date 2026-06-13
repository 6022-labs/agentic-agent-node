package use_cases

import (
	"context"

	identity_repositories "github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	runtime_common_settings "github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	"github.com/google/uuid"
)

type UpdatePaymentRule struct {
	*AbstractPaymentUseCase
}

func NewUpdatePaymentRule(
	agentRepository identity_repositories.AgentRepositoryInterface,
	paymentPolicyRepository repositories.PaymentPolicyRepositoryInterface,
	paymentRuleRepository repositories.PaymentRuleRepositoryInterface,
	paymentMethodStrategies payment_interfaces.PaymentMethodStrategiesInterface,
	blockchainSettings *runtime_common_settings.BlockchainSettings,
	token6022Settings *runtime_common_settings.Token6022Settings,
) *UpdatePaymentRule {
	return &UpdatePaymentRule{
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

// Execute replaces the rule's fields; identity, position and policy binding are
// preserved.
func (u *UpdatePaymentRule) Execute(ctx context.Context, id uuid.UUID, request *requests.SavePaymentRuleRequest) (*responses.PaymentRuleResponse, error) {
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

	rule, err := u.paymentRuleRepository.GetById(ctx, id)
	if err != nil {
		return nil, errors.NewInternalError("update_failed", err.Error())
	}
	if rule == nil {
		return nil, errors.NewNotFoundError("rule_not_found", "")
	}

	rule.Name = request.Name
	rule.Price = request.Price
	rule.Method = enums.PaymentMethod(request.Method)

	updated, err := u.paymentRuleRepository.Update(ctx, rule)
	if err != nil {
		return nil, errors.NewInternalError("update_failed", err.Error())
	}
	if err := strategy.SaveConfig(ctx, updated.Id, request.Config); err != nil {
		return nil, errors.NewInternalError("update_failed", err.Error())
	}
	return responses.ToPaymentRuleResponse(updated, request.Config), nil
}
