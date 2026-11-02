package services

import (
	"context"
	"math/big"
	"strings"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	runtime_common_interfaces "github.com/6022-labs/agentic-agent-node/src/runtime_common/services/interfaces"
	"github.com/google/uuid"
)

// X402PaymentSpendPolicyEvaluator denies by default: an unlisted payee or a token without a rule is refused.
type X402PaymentSpendPolicyEvaluator struct {
	paymentSpendPolicyProvider           payment_interfaces.PaymentSpendPolicyProviderInterface
	paymentSpendRuleRepository           repositories.PaymentSpendRuleRepositoryInterface
	paymentSpendRuleX402ConfigRepository repositories.PaymentSpendRuleX402ConfigRepositoryInterface
	paymentSpendRuleWindowRepository     repositories.PaymentSpendRuleWindowRepositoryInterface
	unitOfWorkFactory                    runtime_common_interfaces.UnitOfWorkFactoryInterface
}

func NewX402PaymentSpendPolicyEvaluator(
	paymentSpendPolicyProvider payment_interfaces.PaymentSpendPolicyProviderInterface,
	paymentSpendRuleRepository repositories.PaymentSpendRuleRepositoryInterface,
	paymentSpendRuleX402ConfigRepository repositories.PaymentSpendRuleX402ConfigRepositoryInterface,
	paymentSpendRuleWindowRepository repositories.PaymentSpendRuleWindowRepositoryInterface,
	unitOfWorkFactory runtime_common_interfaces.UnitOfWorkFactoryInterface,
) *X402PaymentSpendPolicyEvaluator {
	return &X402PaymentSpendPolicyEvaluator{
		paymentSpendPolicyProvider:           paymentSpendPolicyProvider,
		paymentSpendRuleRepository:           paymentSpendRuleRepository,
		paymentSpendRuleX402ConfigRepository: paymentSpendRuleX402ConfigRepository,
		paymentSpendRuleWindowRepository:     paymentSpendRuleWindowRepository,
		unitOfWorkFactory:                    unitOfWorkFactory,
	}
}

func (e *X402PaymentSpendPolicyEvaluator) Authorize(
	ctx context.Context,
	quote *payment_interfaces.X402PaymentQuote,
	amount *big.Int,
) (*payment_interfaces.X402PaymentSpendReservation, error) {
	policy, err := e.paymentSpendPolicyProvider.Ensure(ctx)
	if err != nil {
		return nil, err
	}
	if !policy.AllowsPayee(quote.PayTo) {
		return nil, nil
	}

	// Deciding and recording share one transaction, so concurrent signatures cannot overrun a window.
	unitOfWork, err := e.unitOfWorkFactory.Start(ctx)
	if err != nil {
		return nil, errors.NewInternalError("spend_policy_failed", err.Error())
	}
	defer func() { _ = unitOfWork.Rollback() }()

	reservation, err := e.recordSpend(unitOfWork.Context(), policy, quote, amount)
	if err != nil {
		return nil, errors.NewInternalError("spend_policy_failed", err.Error())
	}
	// A refusal wrote nothing, so there is nothing worth keeping.
	if reservation == nil {
		return nil, nil
	}
	if err := unitOfWork.Commit(); err != nil {
		return nil, errors.NewInternalError("spend_policy_failed", err.Error())
	}
	return reservation, nil
}

func (e *X402PaymentSpendPolicyEvaluator) Release(
	ctx context.Context,
	reservation *payment_interfaces.X402PaymentSpendReservation,
) error {
	if len(reservation.WindowStarts) == 0 {
		return nil
	}

	unitOfWork, err := e.unitOfWorkFactory.Start(ctx)
	if err != nil {
		return errors.NewInternalError("spend_policy_failed", err.Error())
	}
	defer func() { _ = unitOfWork.Rollback() }()

	windows, err := e.paymentSpendRuleWindowRepository.ListByRuleIdForUpdate(unitOfWork.Context(), reservation.RuleId)
	if err != nil {
		return errors.NewInternalError("spend_policy_failed", err.Error())
	}
	for i := range windows {
		startedAt, charged := reservation.WindowStarts[windows[i].Id]
		if !charged {
			continue
		}
		if err := windows[i].Release(reservation.Amount, startedAt); err != nil {
			return errors.NewInternalError("spend_policy_failed", err.Error())
		}
		if _, err := e.paymentSpendRuleWindowRepository.Update(unitOfWork.Context(), &windows[i]); err != nil {
			return errors.NewInternalError("spend_policy_failed", err.Error())
		}
	}
	if err := unitOfWork.Commit(); err != nil {
		return errors.NewInternalError("spend_policy_failed", err.Error())
	}
	return nil
}

// recordSpend checks the token's rule and records the amount on each of its windows, all or none.
func (e *X402PaymentSpendPolicyEvaluator) recordSpend(
	ctx context.Context,
	policy *models.PaymentSpendPolicy,
	quote *payment_interfaces.X402PaymentQuote,
	amount *big.Int,
) (*payment_interfaces.X402PaymentSpendReservation, error) {
	// Stored lowercased by the strategy, so the lookup must match that form.
	config, err := e.paymentSpendRuleX402ConfigRepository.GetByPolicyIdAndAsset(
		ctx,
		policy.Id,
		strings.TrimSpace(quote.Network),
		strings.ToLower(strings.TrimSpace(quote.Asset)),
	)
	if err != nil {
		return nil, err
	}
	if config == nil || !config.AcceptsAssetTransferMethod(quote.AssetTransferMethod.String()) {
		return nil, nil
	}

	rule, err := e.paymentSpendRuleRepository.GetById(ctx, config.PaymentSpendRuleId)
	if err != nil {
		return nil, err
	}
	if rule == nil || rule.Method != enums.PaymentMethodX402 {
		return nil, nil
	}
	covered, err := rule.CoversPayment(amount)
	if err != nil || !covered {
		return nil, err
	}

	windows, err := e.paymentSpendRuleWindowRepository.ListByRuleIdForUpdate(ctx, rule.Id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	reservation := &payment_interfaces.X402PaymentSpendReservation{
		RuleId:       rule.Id,
		Amount:       new(big.Int).Set(amount),
		WindowStarts: make(map[uuid.UUID]time.Time, len(windows)),
	}
	for i := range windows {
		recorded, err := windows[i].RecordSpend(amount, now)
		if err != nil || !recorded {
			return nil, err
		}
		reservation.WindowStarts[windows[i].Id] = *windows[i].StartedAt
	}
	for i := range windows {
		if _, err := e.paymentSpendRuleWindowRepository.Update(ctx, &windows[i]); err != nil {
			return nil, err
		}
	}
	return reservation, nil
}
