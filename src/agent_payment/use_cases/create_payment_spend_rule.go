package use_cases

import (
	"context"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type CreatePaymentSpendRule struct {
	*AbstractPaymentSpendUseCase
}

func NewCreatePaymentSpendRule(
	paymentSpendPolicyProvider payment_interfaces.PaymentSpendPolicyProviderInterface,
	paymentSpendRuleRepository repositories.PaymentSpendRuleRepositoryInterface,
	paymentSpendMethodStrategies payment_interfaces.PaymentSpendMethodStrategiesInterface,
) *CreatePaymentSpendRule {
	return &CreatePaymentSpendRule{
		AbstractPaymentSpendUseCase: NewAbstractPaymentSpendUseCase(
			paymentSpendPolicyProvider,
			paymentSpendRuleRepository,
			paymentSpendMethodStrategies,
		),
	}
}

func (u *CreatePaymentSpendRule) Execute(
	ctx context.Context,
	request *requests.SavePaymentSpendRuleRequest,
) (*responses.PaymentSpendRuleResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}
	strategy, ok := u.paymentSpendMethodStrategies.For(enums.PaymentMethod(request.Method))
	if !ok {
		return nil, errors.NewValidationError("method", "unsupported payment method")
	}
	if err := strategy.Validate(ctx, request); err != nil {
		return nil, err
	}

	policy, err := u.paymentSpendPolicyProvider.Ensure(ctx)
	if err != nil {
		return nil, err
	}

	rule, err := u.paymentSpendRuleRepository.Create(ctx, u.toPaymentSpendRule(policy, request))
	if err != nil {
		return nil, errors.NewInternalError("create_failed", err.Error())
	}
	// Persist the matcher after the rule (it needs the rule id).
	if err := strategy.SaveConfig(ctx, rule.Id, request.Config); err != nil {
		_ = u.paymentSpendRuleRepository.DeleteById(ctx, rule.Id)
		return nil, errors.NewInternalError("create_failed", err.Error())
	}
	return responses.ToPaymentSpendRuleResponse(rule, request.Config, time.Now()), nil
}

func (u *CreatePaymentSpendRule) toPaymentSpendRule(
	policy *models.PaymentSpendPolicy,
	request *requests.SavePaymentSpendRuleRequest,
) *models.PaymentSpendRule {
	return &models.PaymentSpendRule{
		PolicyId:            policy.Id,
		Name:                request.Name,
		MaxAmountPerPayment: request.MaxAmountPerPayment,
		WindowSeconds:       request.WindowSeconds,
		MaxAmountPerWindow:  request.MaxAmountPerWindow,
		Method:              enums.PaymentMethod(request.Method),
	}
}
