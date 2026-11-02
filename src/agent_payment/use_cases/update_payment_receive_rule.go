package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	runtime_common_interfaces "github.com/6022-labs/agentic-agent-node/src/runtime_common/services/interfaces"
)

type UpdatePaymentReceiveRule struct {
	*AbstractPaymentUseCase
}

func NewUpdatePaymentReceiveRule(
	paymentReceivePolicyProvider payment_interfaces.PaymentReceivePolicyProviderInterface,
	paymentReceiveRuleRepository repositories.PaymentReceiveRuleRepositoryInterface,
	paymentReceiveRulePriceRepository repositories.PaymentReceiveRulePriceRepositoryInterface,
	paymentReceiveMethodStrategies payment_interfaces.PaymentReceiveMethodStrategiesInterface,
	unitOfWorkFactory runtime_common_interfaces.UnitOfWorkFactoryInterface,
) *UpdatePaymentReceiveRule {
	return &UpdatePaymentReceiveRule{
		AbstractPaymentUseCase: NewAbstractPaymentUseCase(
			paymentReceivePolicyProvider,
			paymentReceiveRuleRepository,
			paymentReceiveRulePriceRepository,
			paymentReceiveMethodStrategies,
			unitOfWorkFactory,
		),
	}
}

// Execute replaces the rule's token and prices, keeping its id.
func (u *UpdatePaymentReceiveRule) Execute(
	ctx context.Context,
	request *requests.UpdatePaymentReceiveRuleRequest,
) (*responses.PaymentReceiveRuleResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}
	strategy, ok := u.paymentReceiveMethodStrategies.For(enums.PaymentMethod(request.Method))
	if !ok {
		return nil, errors.NewValidationError("method", "unsupported payment method")
	}
	if err := strategy.Validate(ctx, &request.SavePaymentReceiveRuleRequest); err != nil {
		return nil, err
	}

	rule, err := u.paymentReceiveRuleRepository.GetById(ctx, request.Id)
	if err != nil {
		return nil, errors.NewInternalError("update_failed", err.Error())
	}
	if rule == nil {
		return nil, errors.NewNotFoundError("rule_not_found", "")
	}
	if err := u.claimAsset(ctx, strategy, rule.PolicyId, rule.Id, request.Config); err != nil {
		return nil, err
	}

	unitOfWork, err := u.unitOfWorkFactory.Start(ctx)
	if err != nil {
		return nil, errors.NewInternalError("update_failed", err.Error())
	}
	defer func() { _ = unitOfWork.Rollback() }()

	rule.Name = request.Name
	rule.Method = enums.PaymentMethod(request.Method)
	updated, err := u.paymentReceiveRuleRepository.Update(unitOfWork.Context(), rule)
	if err != nil {
		return nil, errors.NewInternalError("update_failed", err.Error())
	}
	if err := strategy.SaveConfig(unitOfWork.Context(), rule.PolicyId, rule.Id, request.Config); err != nil {
		return nil, errors.NewInternalError("update_failed", err.Error())
	}
	// Prices carry no state, so they are replaced wholesale.
	if err := u.paymentReceiveRulePriceRepository.DeleteByRuleId(unitOfWork.Context(), rule.Id); err != nil {
		return nil, errors.NewInternalError("update_failed", err.Error())
	}
	prices, err := u.createPrices(unitOfWork.Context(), rule.Id, request.Prices)
	if err != nil {
		return nil, errors.NewInternalError("update_failed", err.Error())
	}
	if err := unitOfWork.Commit(); err != nil {
		return nil, errors.NewInternalError("update_failed", err.Error())
	}
	return responses.ToPaymentReceiveRuleResponse(updated, prices, request.Config), nil
}
