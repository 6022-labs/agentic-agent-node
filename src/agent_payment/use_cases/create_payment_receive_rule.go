package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	runtime_common_interfaces "github.com/6022-labs/agentic-agent-node/src/runtime_common/services/interfaces"
	"github.com/google/uuid"
)

type CreatePaymentReceiveRule struct {
	*AbstractPaymentUseCase
}

func NewCreatePaymentReceiveRule(
	paymentReceivePolicyProvider payment_interfaces.PaymentReceivePolicyProviderInterface,
	paymentReceiveRuleRepository repositories.PaymentReceiveRuleRepositoryInterface,
	paymentReceiveRulePriceRepository repositories.PaymentReceiveRulePriceRepositoryInterface,
	paymentReceiveMethodStrategies payment_interfaces.PaymentReceiveMethodStrategiesInterface,
	unitOfWorkFactory runtime_common_interfaces.UnitOfWorkFactoryInterface,
) *CreatePaymentReceiveRule {
	return &CreatePaymentReceiveRule{
		AbstractPaymentUseCase: NewAbstractPaymentUseCase(
			paymentReceivePolicyProvider,
			paymentReceiveRuleRepository,
			paymentReceiveRulePriceRepository,
			paymentReceiveMethodStrategies,
			unitOfWorkFactory,
		),
	}
}

func (u *CreatePaymentReceiveRule) Execute(
	ctx context.Context,
	request *requests.SavePaymentReceiveRuleRequest,
) (*responses.PaymentReceiveRuleResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}
	strategy, ok := u.paymentReceiveMethodStrategies.For(enums.PaymentMethod(request.Method))
	if !ok {
		return nil, errors.NewValidationError("method", "unsupported payment method")
	}
	if err := strategy.Validate(ctx, request); err != nil {
		return nil, err
	}

	policy, _, err := u.ensurePaymentReceivePolicy(ctx)
	if err != nil {
		return nil, err
	}
	if err := u.claimAsset(ctx, strategy, policy.Id, uuid.Nil, request.Config); err != nil {
		return nil, err
	}

	unitOfWork, err := u.unitOfWorkFactory.Start(ctx)
	if err != nil {
		return nil, errors.NewInternalError("create_failed", err)
	}
	defer func() { _ = unitOfWork.Rollback() }()

	rule, err := u.paymentReceiveRuleRepository.Create(unitOfWork.Context(), &models.PaymentReceiveRule{
		PolicyId: policy.Id,
		Name:     request.Name,
		Method:   enums.PaymentMethod(request.Method),
	})
	if err != nil {
		return nil, errors.NewInternalError("create_failed", err)
	}
	if err := strategy.SaveConfig(unitOfWork.Context(), policy.Id, rule.Id, request.Config); err != nil {
		return nil, errors.NewInternalError("create_failed", err)
	}
	prices, err := u.createPrices(unitOfWork.Context(), rule.Id, request.Prices)
	if err != nil {
		return nil, errors.NewInternalError("create_failed", err)
	}
	if err := unitOfWork.Commit(); err != nil {
		return nil, errors.NewInternalError("create_failed", err)
	}
	return responses.ToPaymentReceiveRuleResponse(rule, prices, request.Config), nil
}
