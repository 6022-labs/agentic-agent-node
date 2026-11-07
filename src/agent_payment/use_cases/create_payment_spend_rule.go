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
	runtime_common_interfaces "github.com/6022-labs/agentic-agent-node/src/runtime_common/services/interfaces"
	"github.com/google/uuid"
)

type CreatePaymentSpendRule struct {
	*AbstractPaymentSpendUseCase
}

func NewCreatePaymentSpendRule(
	paymentSpendPolicyProvider payment_interfaces.PaymentSpendPolicyProviderInterface,
	paymentSpendRuleRepository repositories.PaymentSpendRuleRepositoryInterface,
	paymentSpendRuleWindowRepository repositories.PaymentSpendRuleWindowRepositoryInterface,
	paymentSpendMethodStrategies payment_interfaces.PaymentSpendMethodStrategiesInterface,
	unitOfWorkFactory runtime_common_interfaces.UnitOfWorkFactoryInterface,
) *CreatePaymentSpendRule {
	return &CreatePaymentSpendRule{
		AbstractPaymentSpendUseCase: NewAbstractPaymentSpendUseCase(
			paymentSpendPolicyProvider,
			paymentSpendRuleRepository,
			paymentSpendRuleWindowRepository,
			paymentSpendMethodStrategies,
			unitOfWorkFactory,
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
	if _, err := u.claimAsset(ctx, strategy, policy.Id, uuid.Nil, request.Config); err != nil {
		return nil, err
	}

	unitOfWork, err := u.unitOfWorkFactory.Start(ctx)
	if err != nil {
		return nil, errors.NewInternalError("create_failed", err)
	}
	defer func() { _ = unitOfWork.Rollback() }()

	rule, err := u.paymentSpendRuleRepository.Create(unitOfWork.Context(), &models.PaymentSpendRule{
		PolicyId:            policy.Id,
		Name:                request.Name,
		Method:              enums.PaymentMethod(request.Method),
		MaxAmountPerPayment: request.MaxAmountPerPayment,
	})
	if err != nil {
		return nil, errors.NewInternalError("create_failed", err)
	}
	if err := strategy.SaveConfig(unitOfWork.Context(), policy.Id, rule.Id, request.Config); err != nil {
		return nil, errors.NewInternalError("create_failed", err)
	}

	windows := make([]models.PaymentSpendRuleWindow, 0, len(request.Windows))
	for i := range request.Windows {
		window, err := u.paymentSpendRuleWindowRepository.Create(unitOfWork.Context(), &models.PaymentSpendRuleWindow{
			PaymentSpendRuleId: rule.Id,
			WindowSeconds:      request.Windows[i].WindowSeconds,
			MaxAmount:          request.Windows[i].MaxAmount,
			Spent:              "0",
		})
		if err != nil {
			return nil, errors.NewInternalError("create_failed", err)
		}
		windows = append(windows, *window)
	}

	if err := unitOfWork.Commit(); err != nil {
		return nil, errors.NewInternalError("create_failed", err)
	}
	return responses.ToPaymentSpendRuleResponse(rule, windows, request.Config, time.Now()), nil
}
