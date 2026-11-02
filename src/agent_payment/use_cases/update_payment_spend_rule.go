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

type UpdatePaymentSpendRule struct {
	*AbstractPaymentSpendUseCase
}

func NewUpdatePaymentSpendRule(
	paymentSpendPolicyProvider payment_interfaces.PaymentSpendPolicyProviderInterface,
	paymentSpendRuleRepository repositories.PaymentSpendRuleRepositoryInterface,
	paymentSpendRuleWindowRepository repositories.PaymentSpendRuleWindowRepositoryInterface,
	paymentSpendMethodStrategies payment_interfaces.PaymentSpendMethodStrategiesInterface,
	unitOfWorkFactory runtime_common_interfaces.UnitOfWorkFactoryInterface,
) *UpdatePaymentSpendRule {
	return &UpdatePaymentSpendRule{
		AbstractPaymentSpendUseCase: NewAbstractPaymentSpendUseCase(
			paymentSpendPolicyProvider,
			paymentSpendRuleRepository,
			paymentSpendRuleWindowRepository,
			paymentSpendMethodStrategies,
			unitOfWorkFactory,
		),
	}
}

// Execute replaces the rule; a window keeps its running spend while its period and token stay the same.
func (u *UpdatePaymentSpendRule) Execute(
	ctx context.Context,
	request *requests.UpdatePaymentSpendRuleRequest,
) (*responses.PaymentSpendRuleResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}
	strategy, ok := u.paymentSpendMethodStrategies.For(enums.PaymentMethod(request.Method))
	if !ok {
		return nil, errors.NewValidationError("method", "unsupported payment method")
	}
	if err := strategy.Validate(ctx, &request.SavePaymentSpendRuleRequest); err != nil {
		return nil, err
	}

	rule, err := u.paymentSpendRuleRepository.GetById(ctx, request.Id)
	if err != nil {
		return nil, errors.NewInternalError("update_failed", err.Error())
	}
	if rule == nil {
		return nil, errors.NewNotFoundError("rule_not_found", "")
	}
	sameAsset, err := u.claimAsset(ctx, strategy, rule.PolicyId, rule.Id, request.Config)
	if err != nil {
		return nil, err
	}

	unitOfWork, err := u.unitOfWorkFactory.Start(ctx)
	if err != nil {
		return nil, errors.NewInternalError("update_failed", err.Error())
	}
	defer func() { _ = unitOfWork.Rollback() }()

	rule.Name = request.Name
	rule.Method = enums.PaymentMethod(request.Method)
	rule.MaxAmountPerPayment = request.MaxAmountPerPayment
	updated, err := u.paymentSpendRuleRepository.Update(unitOfWork.Context(), rule)
	if err != nil {
		return nil, errors.NewInternalError("update_failed", err.Error())
	}
	if err := strategy.SaveConfig(unitOfWork.Context(), rule.PolicyId, rule.Id, request.Config); err != nil {
		return nil, errors.NewInternalError("update_failed", err.Error())
	}
	windows, err := u.replaceWindows(unitOfWork.Context(), rule.Id, request.Windows, sameAsset)
	if err != nil {
		return nil, errors.NewInternalError("update_failed", err.Error())
	}

	if err := unitOfWork.Commit(); err != nil {
		return nil, errors.NewInternalError("update_failed", err.Error())
	}
	return responses.ToPaymentSpendRuleResponse(updated, windows, request.Config, time.Now()), nil
}

// replaceWindows matches windows by period: kept ones keep their spend unless the token changed.
func (u *UpdatePaymentSpendRule) replaceWindows(
	ctx context.Context,
	ruleId uuid.UUID,
	requested []requests.SavePaymentSpendRuleWindowRequest,
	sameAsset bool,
) ([]models.PaymentSpendRuleWindow, error) {
	// Locked so a payment recording spend meanwhile is not overwritten by the stale row.
	existing, err := u.paymentSpendRuleWindowRepository.ListByRuleIdForUpdate(ctx, ruleId)
	if err != nil {
		return nil, err
	}
	bySeconds := make(map[int]models.PaymentSpendRuleWindow, len(existing))
	for i := range existing {
		bySeconds[existing[i].WindowSeconds] = existing[i]
	}

	windows := make([]models.PaymentSpendRuleWindow, 0, len(requested))
	for i := range requested {
		window, kept := bySeconds[requested[i].WindowSeconds]
		delete(bySeconds, requested[i].WindowSeconds)
		if !kept {
			window = models.PaymentSpendRuleWindow{
				PaymentSpendRuleId: ruleId,
				WindowSeconds:      requested[i].WindowSeconds,
				Spent:              "0",
			}
		}
		window.MaxAmount = requested[i].MaxAmount
		if !sameAsset {
			window.Reset()
		}

		var saved *models.PaymentSpendRuleWindow
		if kept {
			saved, err = u.paymentSpendRuleWindowRepository.Update(ctx, &window)
		} else {
			saved, err = u.paymentSpendRuleWindowRepository.Create(ctx, &window)
		}
		if err != nil {
			return nil, err
		}
		windows = append(windows, *saved)
	}

	for _, removed := range bySeconds {
		if err := u.paymentSpendRuleWindowRepository.DeleteById(ctx, removed.Id); err != nil {
			return nil, err
		}
	}
	return windows, nil
}
